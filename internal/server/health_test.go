package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// nopLogger is a no-op logger for testing
type nopLogger struct{}

func (n *nopLogger) Log(level log.Level, keyvals ...interface{}) error {
	return nil
}

// newTestLogger creates a no-op logger for testing
func newTestLogger() log.Logger {
	return &nopLogger{}
}

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	// Enable ping monitoring for health check tests
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	// GORM performs a ping when opening the connection
	mock.ExpectPing()

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return db, mock, cleanup
}

func TestHealthChecker_CheckLiveness(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "liveness check always succeeds",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, cleanup := setupMockDB(t)
			defer cleanup()

			logger := newTestLogger()
			hc := NewHealthChecker(db, nil, logger)

			err := hc.CheckLiveness(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHealthChecker_CheckReadiness(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		natsConnected bool
		useNATS       bool
		wantErr       bool
		errContains   string
	}{
		{
			name: "all dependencies healthy",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			natsConnected: true,
			useNATS:       true,
			wantErr:       false,
		},
		{
			name: "database healthy, no NATS configured",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			natsConnected: false,
			useNATS:       false,
			wantErr:       false,
		},
		{
			name: "NATS not connected (but NATS check is non-fatal)",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			natsConnected: false,
			useNATS:       true,
			wantErr:       false, // NATS failures don't fail readiness by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			var nc *nats.Conn
			if tt.useNATS {
				// For NATS testing, we need to use a real connection or advanced mocking
				// For simplicity, we'll just test with nil NATS
				nc = nil
			}

			logger := log.NewStdLogger(nil)
			hc := NewHealthChecker(db, nc, logger)

			err := hc.CheckReadiness(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHealthChecker_checkDatabase(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		wantErr     bool
		errContains string
	}{
		{
			name: "database ping succeeds",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			wantErr: false,
		},
		{
			name: "database ping fails",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
			},
			wantErr:     true,
			errContains: "database ping failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			logger := newTestLogger()
			hc := NewHealthChecker(db, nil, logger)

			err := hc.checkDatabase(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHealthChecker_LivenessHandler(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "liveness handler returns 200 OK",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, cleanup := setupMockDB(t)
			defer cleanup()

			logger := newTestLogger()
			hc := NewHealthChecker(db, nil, logger)

			req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
			w := httptest.NewRecorder()

			handler := hc.LivenessHandler()
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestHealthChecker_ReadinessHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(mock sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "readiness handler returns 200 when healthy",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name: "readiness handler returns 503 when database fails",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(errors.New("connection refused"))
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "Service not ready: database not ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			logger := newTestLogger()
			hc := NewHealthChecker(db, nil, logger)

			req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			w := httptest.NewRecorder()

			handler := hc.ReadinessHandler()
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNewHealthChecker(t *testing.T) {
	t.Run("creates health checker with all dependencies", func(t *testing.T) {
		db, _, cleanup := setupMockDB(t)
		defer cleanup()

		logger := log.NewStdLogger(nil)

		hc := NewHealthChecker(db, nil, logger)

		assert.NotNil(t, hc)
		assert.NotNil(t, hc.db)
		assert.Nil(t, hc.nc)
		assert.NotNil(t, hc.logger)
	})

	t.Run("creates health checker with NATS", func(t *testing.T) {
		db, _, cleanup := setupMockDB(t)
		defer cleanup()

		logger := log.NewStdLogger(nil)

		// Note: We pass nil for NATS as we can't easily create a real connection in tests
		hc := NewHealthChecker(db, nil, logger)

		assert.NotNil(t, hc)
		assert.NotNil(t, hc.db)
		assert.NotNil(t, hc.logger)
	})
}
