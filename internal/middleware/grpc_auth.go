package middleware

import (
	"context"
	"dispatch-and-delivery/internal/models"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// contextKey is a private type to avoid collisions in context keys.
type contextKey string

const (
	UserIDContextKey   contextKey = "userID"
	UserRoleContextKey contextKey = "userRole"
)

// AuthInterceptor is a gRPC server-side interceptor for JWT authentication and authorization.
func AuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		// Extract token from incoming gRPC metadata (the equivalent of HTTP headers)
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		// The token is expected to be in the format "Bearer <token>"
		authHeader := values[0]
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token format is invalid")
		}
		tokenString := parts[1]

		// Parse and validate the token
		claims := &models.JwtCustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return nil, status.Errorf(codes.Unauthenticated, "token is invalid: %v", err)
		}

		// --- Authentication successful ---
		// Inject user information into the context for downstream handlers to use.
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleContextKey, claims.Role)

		// --- Authorization (Role Check) ---
		// This is an example of how you could protect specific methods.
		// A more robust solution might use a map of methods to required roles.
		if info.FullMethod == "/admin.AdminService/SomeAdminAction" {
			if claims.Role != "ADMIN" {
				return nil, status.Errorf(codes.PermissionDenied, "you do not have permission to access this resource")
			}
		}

		// Call the actual RPC handler with the enriched context
		return handler(ctx, req)
	}
}
