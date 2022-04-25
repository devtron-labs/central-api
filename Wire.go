//+build wireinject

package main

import (
	"github.com/devtron-labs/central-api/api"
	util "github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/internal/logger"
	"github.com/devtron-labs/central-api/pkg"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		NewApp,
		api.NewMuxRouter,
		logger.NewSugardLogger,
		util.NewGitHubClient,
		util.NewReleaseCache,
		//logger.NewHttpClient,
		api.NewRestHandlerImpl,
		wire.Bind(new(api.RestHandler), new(*api.RestHandlerImpl)),
		pkg.NewReleaseNoteServiceImpl,
		wire.Bind(new(pkg.ReleaseNoteService), new(*pkg.ReleaseNoteServiceImpl)),
		pkg.NewWebhookSecretValidatorImpl,
		wire.Bind(new(pkg.WebhookSecretValidator), new(*pkg.WebhookSecretValidatorImpl)),

	)
	return &App{}, nil
}
