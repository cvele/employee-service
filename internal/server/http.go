package server

import (
	employee "github.com/cvele/employee-service/api/employee/v1"
	"github.com/cvele/employee-service/internal/conf"
	"github.com/cvele/employee-service/internal/observability"
	"github.com/cvele/employee-service/internal/server/middleware"
	"github.com/cvele/employee-service/internal/service"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	auth *conf.Auth,
	obs *observability.Observability,
	employeeSvc *service.EmployeeService,
	healthChecker *HealthChecker,
	logger log.Logger,
) *http.Server {
	// Get JWT secret from environment variable or config
	jwtSecret := auth.JwtSecret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not configured")
	}

	// Build middleware chain
	middlewares := []kratosMiddleware.Middleware{
		recovery.Recovery(),
	}

	// Add observability middleware (tracing, logging, metrics)
	middlewares = append(middlewares, obs.ServerMiddleware()...)

	// Add business middleware
	middlewares = append(middlewares,
		middleware.ProtoValidate(),
		middleware.JWTAuth(jwtSecret),
	)

	var opts = []http.ServerOption{
		http.Middleware(middlewares...),
	}

	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)

	// Register service
	employee.RegisterEmployeeServiceHTTPServer(srv, employeeSvc)

	// Register metrics endpoint (no auth required)
	srv.Handle("/metrics", observability.MetricsHandler())

	// Register health check endpoints (no auth required)
	srv.HandleFunc("/health/live", healthChecker.LivenessHandler())
	srv.HandleFunc("/health/ready", healthChecker.ReadinessHandler())

	return srv
}
