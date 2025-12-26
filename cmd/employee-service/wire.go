//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"employee-service/internal/biz"
	"employee-service/internal/conf"
	"employee-service/internal/data"
	"employee-service/internal/observability"
	"employee-service/internal/server"
	"employee-service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(
	serverConf *conf.Server,
	dataConf *conf.Data,
	authConf *conf.Auth,
	obsConf *conf.Observability,
	environment string,
	serviceName observability.ServiceName,
	version observability.ServiceVersion,
	logger log.Logger,
) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		observability.ProviderSet,
		observability.NewServiceInfo,
		newApp,
	))
}
