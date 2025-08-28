package utils

import (
	"context"
	"dispatch-and-delivery/internal/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

// GetUserIDFromContext gets userID from context (set by interceptor)
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok || userID == "" {
		return "", status.Error(codes.Unauthenticated, "user ID not found in context")
	}
	return userID, nil
}

// GetPaginationParams processes pagination parameters from a gRPC request.
// It takes the page and limit provided by the client and returns sanitized values
// with sensible defaults.
func GetPaginationParams(page, limit int32) (int32, int32) {
	// If the client provides a page number less than 1, default to the first page.
	if page < 1 {
		page = DefaultPage
	}

	// If the client provides a limit that is invalid or too large, cap it.
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}

	return page, limit
}
