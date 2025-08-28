package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"dispatch-and-delivery/internal/config"
	"dispatch-and-delivery/internal/database"
	"dispatch-and-delivery/internal/modules/users"
	"dispatch-and-delivery/pkg/email"
	pb "dispatch-and-delivery/pkg/proto/user"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
)

func main() {
	// 1. --- Configuration & Database ---
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	dbPool, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection successful.")

	// 2. --- Dependency Injection (Wiring) ---
	// Initialize Google OAuth Config
	googleOAuthConfig := &oauth2.Config{
		RedirectURL:  cfg.GoogleOAuthRedirectURL,
		ClientID:     cfg.GoogleOAuthClientID,
		ClientSecret: cfg.GoogleOAuthClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	sesSender, err := email.NewSESV2Sender(context.Background(), cfg.AWSRegion, cfg.EmailFromAddress)
	if err != nil {
		log.Fatalf("Failed to create SES sender: %v", err)
	}
	templateManager, err := email.NewTemplateManager()
	if err != nil {
		log.Fatalf("Failed to parse email templates: %v", err)
	}
	// The layers are the same as the monolith: Repository -> Service -> Handler
	userRepo := users.NewRepository(dbPool)
	// For this service, we can pass nil for dependencies it doesn't use (like emailer for now)
	userService := users.NewService(userRepo, sesSender, templateManager, cfg.JWTSecret, cfg.ClientOrigin, googleOAuthConfig)
	userGRPCHandler := users.NewGRPCHandler(userService)

	// 3. --- gRPC Server Setup ---
	// Create a TCP listener on the port defined for this service
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort) // e.g., ":50051"
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register our handler implementation with the gRPC server
	pb.RegisterUserServiceServer(grpcServer, userGRPCHandler)
	log.Printf("gRPC server listening at %v", lis.Addr())

	// 4. --- Start Server with Graceful Shutdown ---
	// Run the server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for an interrupt signal (e.g., Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Initiate a graceful shutdown
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("Server exiting.")
}
