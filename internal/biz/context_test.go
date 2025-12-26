package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTenantID(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
	}{
		{
			name:     "set tenant ID",
			tenantID: "tenant-123",
		},
		{
			name:     "set empty tenant ID",
			tenantID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithTenantID(ctx, tt.tenantID)
			
			// Verify by getting it back
			retrieved, err := GetTenantID(ctx)
			if tt.tenantID == "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.tenantID, retrieved)
			}
		})
	}
}

func TestGetTenantID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() context.Context
		want    string
		wantErr bool
	}{
		{
			name: "get existing tenant ID",
			setup: func() context.Context {
				return WithTenantID(context.Background(), "tenant-123")
			},
			want:    "tenant-123",
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			setup: func() context.Context {
				return context.Background()
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			tenantID, err := GetTenantID(ctx)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrTenantNotFound, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, tenantID)
			}
		})
	}
}

func TestWithUserID(t *testing.T) {
	tests := []struct {
		name   string
		userID string
	}{
		{
			name:   "set user ID",
			userID: "user-456",
		},
		{
			name:   "set empty user ID",
			userID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithUserID(ctx, tt.userID)
			
			// Verify by getting it back
			retrieved, err := GetUserID(ctx)
			if tt.userID == "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, retrieved)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() context.Context
		want    string
		wantErr bool
	}{
		{
			name: "get existing user ID",
			setup: func() context.Context {
				return WithUserID(context.Background(), "user-456")
			},
			want:    "user-456",
			wantErr: false,
		},
		{
			name: "missing user ID",
			setup: func() context.Context {
				return context.Background()
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			userID, err := GetUserID(ctx)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrUserNotFound, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, userID)
			}
		})
	}
}

func TestContextIntegration(t *testing.T) {
	// Test that multiple context values can coexist
	ctx := context.Background()
	ctx = WithTenantID(ctx, "tenant-123")
	ctx = WithUserID(ctx, "user-456")

	tenantID, err := GetTenantID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "tenant-123", tenantID)

	userID, err := GetUserID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "user-456", userID)
}

