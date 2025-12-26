package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/cvele/employee-service/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in JWT token
type JWTClaims struct {
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// JWTAuth creates a JWT authentication middleware
func JWTAuth(jwtSecret string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Extract token from metadata/headers
			token, err := extractToken(ctx)
			if err != nil {
				return nil, errors.Unauthorized("UNAUTHORIZED", "missing or invalid authorization header")
			}

			// Parse and validate token
			claims, err := parseToken(token, jwtSecret)
			if err != nil {
				return nil, errors.Unauthorized("UNAUTHORIZED", fmt.Sprintf("invalid token: %v", err))
			}

			// Validate required claims
			if claims.Subject == "" {
				return nil, errors.Unauthorized("UNAUTHORIZED", "missing sub claim in token")
			}
			if claims.TenantID == "" {
				return nil, errors.Unauthorized("UNAUTHORIZED", "missing tenant_id claim in token")
			}

			// Inject tenant_id and user_id into context
			ctx = biz.WithTenantID(ctx, claims.TenantID)
			ctx = biz.WithUserID(ctx, claims.Subject)

			return handler(ctx, req)
		}
	}
}

// extractToken extracts the JWT token from the context
func extractToken(ctx context.Context) (string, error) {
	// Get transport header from context (works for both HTTP and gRPC)
	if tr, ok := transport.FromServerContext(ctx); ok {
		header := tr.RequestHeader().Get("Authorization")
		if header != "" {
			return parseAuthHeader(header)
		}
		// Try lowercase
		header = tr.RequestHeader().Get("authorization")
		if header != "" {
			return parseAuthHeader(header)
		}
	}

	return "", fmt.Errorf("authorization header not found")
}

// parseAuthHeader parses the Authorization header value
func parseAuthHeader(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid authorization header format")
	}

	scheme := strings.ToLower(parts[0])
	if scheme != "bearer" {
		return "", fmt.Errorf("unsupported authorization scheme: %s", scheme)
	}

	return parts[1], nil
}

// parseToken parses and validates a JWT token
func parseToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
