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
	util "github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/common"
	"github.com/devtron-labs/central-api/pkg"
	"github.com/devtron-labs/central-api/pkg/bean"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
)

type RestHandler interface {
	GetReleases(w http.ResponseWriter, r *http.Request)
	ReleaseWebhookHandler(w http.ResponseWriter, r *http.Request)
	GetModules(w http.ResponseWriter, r *http.Request)
	GetModulesV2(w http.ResponseWriter, r *http.Request)
	GetModuleByName(w http.ResponseWriter, r *http.Request)
	GetDockerfileTemplateMetadata(w http.ResponseWriter, r *http.Request)
	GetBuildpackMetadata(w http.ResponseWriter, r *http.Request)
}

func NewRestHandlerImpl(logger *zap.SugaredLogger, releaseNoteService pkg.ReleaseNoteService,
	webhookSecretValidator pkg.WebhookSecretValidator, client *util.GitHubClient, ciBuildMetadataService pkg.CiBuildMetadataService) *RestHandlerImpl {
	return &RestHandlerImpl{
		logger:                 logger,
		releaseNoteService:     releaseNoteService,
		webhookSecretValidator: webhookSecretValidator,
		client:                 client,
		ciBuildMetadataService: ciBuildMetadataService,
	}
}

type RestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	releaseNoteService     pkg.ReleaseNoteService
	webhookSecretValidator pkg.WebhookSecretValidator
	client                 *util.GitHubClient
	ciBuildMetadataService pkg.CiBuildMetadataService
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "text/html; charset=utf-8")
}

func (impl RestHandlerImpl) WriteJsonResp(w http.ResponseWriter, err error, respBody interface{}, status int) {
	response := common.Response{}
	response.Code = status
	response.Status = http.StatusText(status)
	if err == nil {
		response.Result = respBody
	} else {
		apiErr := &common.ApiError{}
		apiErr.Code = "000" // 000=unknown
		apiErr.InternalMessage = err.Error()
		apiErr.UserMessage = respBody
		response.Errors = []*common.ApiError{apiErr}

	}
	b, err := json.Marshal(response)
	if err != nil {
		impl.logger.Errorw("error in marshaling err object", "err", err)
		status = 500
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func (impl *RestHandlerImpl) GetModules(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("get all modules")
	setupResponse(&w, r)
	modules, err := impl.releaseNoteService.GetModules()
	if err != nil {
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	impl.WriteJsonResp(w, nil, modules, http.StatusOK)
	return
}

func (impl *RestHandlerImpl) GetModulesV2(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("get all modules")
	setupResponse(&w, r)
	modules, err := impl.releaseNoteService.GetModulesV2()
	if err != nil {
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	impl.WriteJsonResp(w, nil, modules, http.StatusOK)
	return
}

func (impl *RestHandlerImpl) GetReleases(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	impl.logger.Debug("get all releases")
	offset := 0
	size := 10
	var err error
	offsetQueryParam := r.URL.Query().Get("offset")
	if len(offsetQueryParam) > 0 {
		offset, err = strconv.Atoi(offsetQueryParam)
		if err != nil {
			impl.WriteJsonResp(w, err, "invalid offset", http.StatusBadRequest)
			return
		}
	}
	sizeQueryParam := r.URL.Query().Get("size")
	if len(sizeQueryParam) > 0 {
		size, err = strconv.Atoi(sizeQueryParam)
		if err != nil {
			impl.WriteJsonResp(w, err, "invalid size", http.StatusBadRequest)
			return
		}
	}
	repo := r.URL.Query().Get("repo")
	repository := bean.Oss
	if len(repo) > 0 {
		repository = bean.Repository(repo)
	}
	//will fetch all the releases from cache and later apply size and offset filter
	response, err := impl.releaseNoteService.GetReleases(repository)
	if err != nil {
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	if size > 0 {
		if offset+size <= len(response) {
			response = response[offset : offset+size]
		} else {
			response = response[offset:]
		}
	}
	if len(response) == 0 {
		response = make([]*common.Release, 0)
	}

	impl.WriteJsonResp(w, nil, response, http.StatusOK)
	return
}

func (impl *RestHandlerImpl) ReleaseWebhookHandler(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("release webhook handler received event")
	// get git host Id and secret from request
	vars := mux.Vars(r)
	secretFromRequest := vars["secret"]
	impl.logger.Debugw("secret found in request", "secret", secretFromRequest)

	// validate signature
	requestBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		impl.logger.Errorw("Cannot read the request body:", "err", err)
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	isValidSig := impl.webhookSecretValidator.ValidateSecret(r, requestBodyBytes)
	impl.logger.Debugw("Secret validation result ", "isValidSig", isValidSig)
	if !isValidSig {
		impl.logger.Error("Signature mismatch")
		impl.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	// validate event type
	eventType := r.Header.Get(impl.client.GitHubConfig.GitHubEventTypeHeader)
	impl.logger.Debugw("webhook event type header", "eventType : ", eventType)
	if len(eventType) == 0 || eventType != bean.EventTypeRelease {
		impl.logger.Errorw("Event type not known ", eventType)
		impl.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	flag, err := impl.releaseNoteService.UpdateReleases(requestBodyBytes)
	if err != nil {
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	impl.WriteJsonResp(w, err, flag, http.StatusOK)
	return
}

func (impl *RestHandlerImpl) GetModuleByName(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("get module meta info by name")
	setupResponse(&w, r)
	vars := mux.Vars(r)
	name := vars["name"]
	module, err := impl.releaseNoteService.GetModuleByName(name)
	if err != nil {
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	impl.WriteJsonResp(w, nil, module, http.StatusOK)
	return
}

func (impl *RestHandlerImpl) GetDockerfileTemplateMetadata(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("get all dockerfile template metadata")
	setupResponse(&w, r)
	dockerfileTemplateMetadata := impl.ciBuildMetadataService.GetDockerfileTemplateMetadata()
	impl.WriteJsonResp(w, nil, dockerfileTemplateMetadata, http.StatusOK)
	return
}
func (impl *RestHandlerImpl) GetBuildpackMetadata(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("get all buildpack metadata")
	setupResponse(&w, r)
	buildpackMetadata := impl.ciBuildMetadataService.GetBuildpackMetadata()
	impl.WriteJsonResp(w, nil, buildpackMetadata, http.StatusOK)
	return
}
