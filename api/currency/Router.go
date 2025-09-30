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

package currency

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Router interface {
	InitCurrencyRoutes(router *mux.Router)
}

// RouterImpl handles routing for currency-related endpoints
type RouterImpl struct {
	logger      *zap.SugaredLogger
	restHandler CurrencyRestHandler
}

// NewRouter creates a new instance of RouterImpl
func NewRouter(logger *zap.SugaredLogger, restHandler CurrencyRestHandler) *RouterImpl {
	return &RouterImpl{
		logger:      logger,
		restHandler: restHandler,
	}
}

// InitCurrencyRoutes initializes all currency-related routes
func (r *RouterImpl) InitCurrencyRoutes(currencyRouter *mux.Router) {
	r.logger.Info("initializing currency routes")

	// Exchange rates endpoints
	currencyRouter.Path("/rates").
		Methods("GET").
		HandlerFunc(r.restHandler.GetExchangeRates)

	r.logger.Info("currency routes initialized successfully")
}
