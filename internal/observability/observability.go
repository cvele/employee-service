package observability

import (
	"context"
	"employee-service/internal/conf"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewObservability)

// ServiceName is the name of the service
type ServiceName string

// ServiceVersion is the version of the service
type ServiceVersion string

// ServiceInfo holds service metadata
type ServiceInfo struct {
	Name    ServiceName
	Version ServiceVersion
}

// NewServiceInfo creates a new ServiceInfo
func NewServiceInfo(name ServiceName, version ServiceVersion) *ServiceInfo {
	return &ServiceInfo{
		Name:    name,
		Version: version,
	}
}

type Observability struct {
	metrics *MetricsProvider
	tracing *TracingProvider
	logger  log.Logger
	conf    *conf.Observability
}

func NewObservability(c *conf.Observability, info *ServiceInfo, logger log.Logger) (*Observability, func(), error) {
	logHelper := log.NewHelper(logger)
	o := &Observability{
		logger: logger,
		conf:   c,
	}

	cleanup := func() {
		if o.tracing != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := o.tracing.Shutdown(ctx); err != nil {
				logHelper.Errorf("failed to shutdown tracing: %v", err)
			}
		}
	}

	// Initialize metrics
	if c.Metrics != nil && c.Metrics.Enabled {
		o.metrics = NewMetricsProvider(c.Metrics.Namespace, c.Metrics.Subsystem)
		logHelper.Info("Metrics enabled")
	}

	// Initialize tracing
	if c.Tracing != nil && c.Tracing.Enabled {
		tp, err := NewTracingProvider(
			info.Name,
			info.Version,
			c.Tracing.Endpoint,
			c.Tracing.SampleRate,
			c.Tracing.Insecure,
			logger,
		)
		if err != nil {
			logHelper.Warnf("failed to initialize tracing (continuing without): %v", err)
		} else {
			o.tracing = tp
			logHelper.Info("Tracing enabled")
		}
	}

	return o, cleanup, nil
}

func (o *Observability) ServerMiddleware() []middleware.Middleware {
	var mws []middleware.Middleware

	// Tracing should be first to capture full span
	if o.tracing != nil {
		mws = append(mws, tracing.Server())
	}

	// Logging middleware
	if o.conf.Logging != nil && o.conf.Logging.Enabled {
		mws = append(mws, logging.Server(o.logger))
	}

	// Metrics middleware
	if o.metrics != nil {
		mws = append(mws, metrics.Server())
	}

	return mws
}
