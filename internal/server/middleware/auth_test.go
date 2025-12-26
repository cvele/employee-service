package middleware

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTransport struct {
	mock.Mock
}

func (m *mockTransport) Kind() transport.Kind {
	return transport.KindGRPC
}

func (m *mockTransport) Endpoint() string {
	return "localhost:8000"
}

func (m *mockTransport) Operation() string {
	return "/employee.v1.EmployeeService/CreateEmployee"
}

func (m *mockTransport) RequestHeader() transport.Header {
	args := m.Called()
	return args.Get(0).(transport.Header)
}

func (m *mockTransport) ReplyHeader() transport.Header {
	args := m.Called()
	return args.Get(0).(transport.Header)
}

type mockHeader struct {
	data map[string][]string
}

func (h *mockHeader) Get(key string) string {
	if vals, ok := h.data[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (h *mockHeader) Set(key string, value string) {
	h.data[key] = []string{value}
}

func (h *mockHeader) Keys() []string {
	keys := make([]string, 0, len(h.data))
	for k := range h.data {
		keys = append(keys, k)
	}
	return keys
}

func (h *mockHeader) Values(key string) []string {
	return h.data[key]
}

func (h *mockHeader) Add(key string, value string) {
	h.data[key] = append(h.data[key], value)
}

func TestJWTAuth(t *testing.T) {
	secretKey := "test-secret-key"

	tests := []struct {
		name      string
		setupCtx  func() context.Context
		wantErr   bool
	}{
		{
			name: "missing authorization header",
			setupCtx: func() context.Context {
				tr := new(mockTransport)
				header := &mockHeader{
					data: map[string][]string{},
				}
				tr.On("RequestHeader").Return(header)
				
				return transport.NewServerContext(context.Background(), tr)
			},
			wantErr: true,
		},
		{
			name: "invalid token format",
			setupCtx: func() context.Context {
				tr := new(mockTransport)
				header := &mockHeader{
					data: map[string][]string{
						"Authorization": {"InvalidFormat"},
					},
				}
				tr.On("RequestHeader").Return(header)
				
				return transport.NewServerContext(context.Background(), tr)
			},
			wantErr: true,
		},
		{
			name: "malformed JWT",
			setupCtx: func() context.Context {
				tr := new(mockTransport)
				header := &mockHeader{
					data: map[string][]string{
						"Authorization": {"Bearer invalid.jwt.token"},
					},
				}
				tr.On("RequestHeader").Return(header)
				
				return transport.NewServerContext(context.Background(), tr)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := JWTAuth(secretKey)
			
			handler := middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			})

			ctx := tt.setupCtx()
			_, err := handler(ctx, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() context.Context
		wantErr bool
	}{
		{
			name: "valid bearer token",
			setup: func() context.Context {
				tr := new(mockTransport)
				header := &mockHeader{
					data: map[string][]string{
						"Authorization": {"Bearer test-token"},
					},
				}
				tr.On("RequestHeader").Return(header)
				return transport.NewServerContext(context.Background(), tr)
			},
			wantErr: false,
		},
		{
			name: "lowercase authorization header",
			setup: func() context.Context {
				tr := new(mockTransport)
				header := &mockHeader{
					data: map[string][]string{
						"authorization": {"Bearer test-token"},
					},
				}
				tr.On("RequestHeader").Return(header)
				return transport.NewServerContext(context.Background(), tr)
			},
			wantErr: false,
		},
		{
			name: "missing header",
			setup: func() context.Context {
				tr := new(mockTransport)
				header := &mockHeader{
					data: map[string][]string{},
				}
				tr.On("RequestHeader").Return(header)
				return transport.NewServerContext(context.Background(), tr)
			},
			wantErr: true,
		},
		{
			name: "no transport context",
			setup: func() context.Context {
				return context.Background()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			token, err := extractToken(ctx)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestParseAuthHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{
			name:    "valid bearer token",
			header:  "Bearer token123",
			want:    "token123",
			wantErr: false,
		},
		{
			name:    "Bearer with uppercase",
			header:  "BEARER token123",
			want:    "token123",
			wantErr: false,
		},
		{
			name:    "missing token",
			header:  "Bearer",
			want:    "",
			wantErr: true,
		},
		{
			name:    "no space",
			header:  "Bearertoken123",
			want:    "",
			wantErr: true,
		},
		{
			name:    "wrong scheme",
			header:  "Basic token123",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty header",
			header:  "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := parseAuthHeader(tt.header)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, token)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	secretKey := "test-secret-key"

	t.Run("malformed token", func(t *testing.T) {
		tokenString := "invalid.jwt.token"
		
		claims, err := parseToken(tokenString, secretKey)
		
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("invalid signature", func(t *testing.T) {
		// Token signed with different secret
		tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsInRlbmFudF9pZCI6InRlbmFudC00NTYiLCJleHAiOjk5OTk5OTk5OTl9.invalid_signature"
		
		claims, err := parseToken(tokenString, secretKey)
		
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("empty token", func(t *testing.T) {
		tokenString := ""
		
		claims, err := parseToken(tokenString, secretKey)
		
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
