package server

import (
	"employee-service/internal/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, ProvideHealthChecker)

// ProvideHealthChecker creates a health checker from the data layer
func ProvideHealthChecker(d *data.Data, logger log.Logger) *HealthChecker {
	return NewHealthChecker(d.GetDB(), d.GetNATS(), logger)
}
