package main

import (
	"flag"
	"os"

	"github.com/cvele/employee-service/internal/conf"
	"github.com/cvele/employee-service/internal/observability"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "employee-service"
	// Version is the version of the compiled software.
	Version = "v0.0.0"
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs/config.yaml", "config path, eg: -conf ./configs/config.yaml")
}

func newApp(logger log.Logger, environment string, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{
			"env": environment,
		}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
			env.NewSource(), // Loads env vars - file's ${VAR:default} will resolve to these
		),
	)
	defer func() {
		_ = c.Close()
	}()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// Create logger with environment context
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"service.env", bc.Environment,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	// Apply log level filter
	if bc.Observability != nil && bc.Observability.Logging != nil && bc.Observability.Logging.Level != "" {
		logger = log.NewFilter(logger, log.FilterLevel(parseLogLevel(bc.Observability.Logging.Level)))
	}

	app, cleanup, err := wireApp(
		bc.Server,
		bc.Data,
		bc.Auth,
		bc.Observability,
		bc.Environment,
		observability.ServiceName(Name),
		observability.ServiceVersion(Version),
		logger,
	)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// parseLogLevel converts string log level to log.Level
func parseLogLevel(level string) log.Level {
	switch level {
	case "debug":
		return log.LevelDebug
	case "info":
		return log.LevelInfo
	case "warn":
		return log.LevelWarn
	case "error":
		return log.LevelError
	default:
		return log.LevelInfo
	}
}
