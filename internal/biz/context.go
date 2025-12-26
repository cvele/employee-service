package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
)

type contextKey string

const (
	tenantIDKey contextKey = "tenant_id"
	userIDKey   contextKey = "user_id"
)

var (
	// ErrTenantNotFound is tenant not found in context.
	ErrTenantNotFound = errors.Unauthorized("TENANT_NOT_FOUND", "tenant not found in context")
	// ErrUserNotFound is user not found in context.
	ErrUserNotFound = errors.Unauthorized("USER_NOT_FOUND", "user not found in context")
)

// GetTenantID extracts tenant_id from context
func GetTenantID(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(tenantIDKey).(string)
	if !ok || tenantID == "" {
		return "", ErrTenantNotFound
	}
	return tenantID, nil
}

// GetUserID extracts user_id from context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", ErrUserNotFound
	}
	return userID, nil
}

// WithTenantID injects tenant_id into context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// WithUserID injects user_id into context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

