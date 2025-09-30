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
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

// CurrencyConfig holds configuration for the currency service
type CurrencyConfig struct {
	// OpenExchangeRates API configuration
	APIKey  string `env:"OPENEXCHANGERATES_API_KEY" envDefault:""`
	BaseURL string `env:"OPENEXCHANGERATES_BASE_URL" envDefault:"https://openexchangerates.org/api"`

	// Cache configuration
	CacheTTLHours int `env:"CURRENCY_CACHE_TTL_HOURS" envDefault:"24"`

	// Default base currency
	BaseCurrency string `env:"CURRENCY_BASE_CURRENCY" envDefault:"USD"`

	// HTTP client timeout
	HTTPTimeoutSeconds int `env:"CURRENCY_HTTP_TIMEOUT_SECONDS" envDefault:"30"`
}

// NewCurrencyConfig creates a new currency configuration from environment variables
func NewCurrencyConfig(logger *zap.SugaredLogger) (*CurrencyConfig, error) {
	cfg := &CurrencyConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error parsing currency config", "err", err)
		return nil, err
	}

	// Validate required configuration
	if cfg.APIKey == "" {
		logger.Warn("OpenExchangeRates API key not provided, service will have limited functionality")
	}

	logger.Infow("currency config loaded",
		"baseURL", cfg.BaseURL,
		"cacheTTL", cfg.CacheTTLHours,
		"baseCurrency", cfg.BaseCurrency,
		"httpTimeout", cfg.HTTPTimeoutSeconds)

	return cfg, nil
}
