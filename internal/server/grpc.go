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
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Server,
	auth *conf.Auth,
	obs *observability.Observability,
	employeeSvc *service.EmployeeService,
	logger log.Logger,
) *grpc.Server {
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

	var opts = []grpc.ServerOption{
		grpc.Middleware(middlewares...),
	}

	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}

	srv := grpc.NewServer(opts...)
	employee.RegisterEmployeeServiceServer(srv, employeeSvc)

	return srv
}
