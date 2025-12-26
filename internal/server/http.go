package server

import (
	employee "employee-service/api/employee/v1"
	"employee-service/internal/conf"
	"employee-service/internal/server/middleware"
	"employee-service/internal/service"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, auth *conf.Auth, employeeSvc *service.EmployeeService, logger log.Logger) *http.Server {
	// Get JWT secret from environment variable or config
	jwtSecret := auth.JwtSecret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not configured")
	}

	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			middleware.ProtoValidate(), // Protovalidate middleware
			middleware.JWTAuth(jwtSecret),
		),
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
	employee.RegisterEmployeeServiceHTTPServer(srv, employeeSvc)
	return srv
}
