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
	"fmt"
	"github.com/devtron-labs/central-api/api/handler"
	"net/http"
	"strings"

	currencyPkg "github.com/devtron-labs/central-api/pkg/currency"
	"go.uber.org/zap"
)

// CurrencyRestHandler defines the interface for currency REST operations
type CurrencyRestHandler interface {
	// GetExchangeRates returns the latest exchange rates
	GetExchangeRates(w http.ResponseWriter, r *http.Request)
}

// CurrencyRestHandlerImpl implements the CurrencyRestHandler interface
type CurrencyRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	currencyService currencyPkg.Service
}

// NewCurrencyRestHandlerImpl creates a new instance of CurrencyRestHandlerImpl
func NewCurrencyRestHandlerImpl(logger *zap.SugaredLogger, currencyService currencyPkg.Service) *CurrencyRestHandlerImpl {
	return &CurrencyRestHandlerImpl{
		logger:          logger,
		currencyService: currencyService,
	}
}

// GetExchangeRates handles GET /currency/rates requests
func (impl *CurrencyRestHandlerImpl) GetExchangeRates(w http.ResponseWriter, r *http.Request) {
	// Get base currency from query parameter (optional)
	base := r.URL.Query().Get("base")
	if base != "" {
		base = strings.ToUpper(base)
		if !currencyPkg.ValidateCurrencyCode(base) {
			impl.logger.Errorw("invalid base currency code", "base", base)
			errMsg := fmt.Sprintf("Invalid base currency code %q", base)
			apiErr := handler.NewApiError(http.StatusBadRequest, errMsg, errMsg)
			handler.WriteJsonResp(w, apiErr, nil, http.StatusBadRequest)
			return
		}
	}

	rates, err := impl.currencyService.GetExchangeRates(r.Context(), base)
	if err != nil {
		impl.logger.Errorw("failed to get exchange rates", "base", base, "err", err)
		handler.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	handler.WriteJsonResp(w, nil, rates, http.StatusOK)
	return
}
