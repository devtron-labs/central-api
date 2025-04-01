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

package api

import (
	"encoding/json"
	"github.com/devtron-labs/central-api/common"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type MuxRouter struct {
	logger      *zap.SugaredLogger
	Router      *mux.Router
	restHandler RestHandler
}

func NewMuxRouter(logger *zap.SugaredLogger, restHandler RestHandler) *MuxRouter {
	return &MuxRouter{logger: logger, Router: mux.NewRouter(), restHandler: restHandler}
}

func (r MuxRouter) Init() {
	r.Router.StrictSlash(true)
	//r.Router.Handle("/metrics", promhttp.Handler())
	r.Router.Path("/health").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		setupResponse(&writer, request)
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = "OK"
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})

	r.Router.Path("/release/notes").HandlerFunc(r.restHandler.GetReleases).Methods("GET")
	r.Router.Path("/release/webhook").HandlerFunc(r.restHandler.ReleaseWebhookHandler).Methods("POST")
	r.Router.Path("/modules").HandlerFunc(r.restHandler.GetModules).Methods("GET")
	r.Router.Path("/dockerfileTemplate").HandlerFunc(r.restHandler.GetDockerfileTemplateMetadata).Methods("GET")
	r.Router.Path("/buildpackMetadata").HandlerFunc(r.restHandler.GetBuildpackMetadata).Methods("GET")
	r.Router.Path("/v2/modules").HandlerFunc(r.restHandler.GetModulesV2).Methods("GET")
	r.Router.Path("/module").
		Queries("name", "{name}").
		HandlerFunc(r.restHandler.GetModuleByName).Methods("GET")
}
