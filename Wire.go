//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/devtron-labs/central-api/api"
	util "github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/internal/logger"
	"github.com/devtron-labs/central-api/pkg"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		logger.NewSugardLogger,
		//sql.PgSqlWireSet,
		//releaseNote.NewReleaseNoteRepositoryImpl,
		//wire.Bind(new(releaseNote.ReleaseNoteRepository), new(*releaseNote.ReleaseNoteRepositoryImpl)),
		blob_storage.NewBlobStorageServiceImpl,
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
		util.NewBlobConfig,

		pkg.NewCiBuildMetadataServiceImpl,
		wire.Bind(new(pkg.CiBuildMetadataService), new(*pkg.CiBuildMetadataServiceImpl)),
	)
	return &App{}, nil
}
