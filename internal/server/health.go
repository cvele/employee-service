package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

// HealthChecker checks the health of service dependencies
type HealthChecker struct {
	db     *gorm.DB
	nc     *nats.Conn
	logger *log.Helper
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *gorm.DB, nc *nats.Conn, logger log.Logger) *HealthChecker {
	return &HealthChecker{
		db:     db,
		nc:     nc,
		logger: log.NewHelper(logger),
	}
}

// CheckLiveness performs a basic liveness check
// This is a simple check that the service is running
func (h *HealthChecker) CheckLiveness(ctx context.Context) error {
	// Liveness probe: just return OK if the service is running
	return nil
}

// CheckReadiness performs a readiness check on all dependencies
// This checks if the service is ready to handle requests
func (h *HealthChecker) CheckReadiness(ctx context.Context) error {
	// Check database connection
	if err := h.checkDatabase(ctx); err != nil {
		h.logger.Warnf("database health check failed: %v", err)
		return fmt.Errorf("database not ready: %w", err)
	}

	// Check NATS connection (only if configured)
	if h.nc != nil {
		if err := h.checkNATS(); err != nil {
			h.logger.Warnf("NATS health check failed: %v", err)
			// Note: NATS is optional, so we just log a warning but don't fail the readiness check
			// If you want NATS to be required for readiness, uncomment the line below:
			// return fmt.Errorf("NATS not ready: %w", err)
		}
	}

	return nil
}

// checkDatabase verifies the database connection is healthy
func (h *HealthChecker) checkDatabase(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// checkNATS verifies the NATS connection is healthy
func (h *HealthChecker) checkNATS() error {
	if h.nc == nil {
		return fmt.Errorf("NATS connection is nil")
	}

	if !h.nc.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	return nil
}

// LivenessHandler returns an HTTP handler for liveness probes
func (h *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.CheckLiveness(r.Context()); err != nil {
			h.logger.Errorf("liveness check failed: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintf(w, "Service not live: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}

// ReadinessHandler returns an HTTP handler for readiness probes
func (h *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.CheckReadiness(r.Context()); err != nil {
			h.logger.Warnf("readiness check failed: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintf(w, "Service not ready: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}
