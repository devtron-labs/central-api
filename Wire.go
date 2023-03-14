//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/central-api/api"
	util "github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/internal/logger"
	"github.com/devtron-labs/central-api/pkg"
	"github.com/devtron-labs/central-api/pkg/releaseNote"
	"github.com/devtron-labs/central-api/pkg/sql"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		logger.NewSugardLogger,
		sql.PgSqlWireSet,
		releaseNote.NewReleaseNoteRepositoryImpl,
		wire.Bind(new(releaseNote.ReleaseNoteRepository), new(*releaseNote.ReleaseNoteRepositoryImpl)),
		NewApp,
		api.NewMuxRouter,
		util.NewGitHubClient,
		//logger.NewHttpClient,
		api.NewRestHandlerImpl,
		wire.Bind(new(api.RestHandler), new(*api.RestHandlerImpl)),
		pkg.NewReleaseNoteServiceImpl,
		wire.Bind(new(pkg.ReleaseNoteService), new(*pkg.ReleaseNoteServiceImpl)),
		pkg.NewWebhookSecretValidatorImpl,
		wire.Bind(new(pkg.WebhookSecretValidator), new(*pkg.WebhookSecretValidatorImpl)),
		util.NewModuleConfig,

		pkg.NewCiBuildMetadataServiceImpl,
		wire.Bind(new(pkg.CiBuildMetadataService), new(*pkg.CiBuildMetadataServiceImpl)),
	)
	return &App{}, nil
}
