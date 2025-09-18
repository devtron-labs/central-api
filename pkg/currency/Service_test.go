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
	"context"
	"encoding/json"
	"github.com/devtron-labs/common-lib/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewCurrencyConfig(t *testing.T) {
	logger, _ := utils.NewSugardLogger()

	config, err := NewCurrencyConfig(logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.BaseURL != "https://openexchangerates.org/api" {
		t.Errorf("Expected default base URL to be https://openexchangerates.org/api, got %s", config.BaseURL)
	}

	if config.BaseCurrency != USD {
		t.Errorf("Expected default base currency to be USD, got %s", config.BaseCurrency)
	}

	if config.CacheTTLHours != 24 {
		t.Errorf("Expected default cache TTL to be 24 hours, got %d", config.CacheTTLHours)
	}
}

func TestCurrencyServiceImpl_CacheRates(t *testing.T) {
	logger := zap.NewNop().Sugar()
	config := &CurrencyConfig{
		APIKey:             "test-key",
		BaseURL:            "https://test.com",
		CacheTTLHours:      24,
		BaseCurrency:       USD,
		HTTPTimeoutSeconds: 30,
	}

	service := NewServiceImpl(config, logger)

	// Create test rates
	rates := &ExchangeRatesResponse{
		Timestamp: time.Now().Unix(),
		Base:      USD,
		Rates: map[string]float64{
			EUR: 0.85,
			GBP: 0.75,
		},
	}

	// Cache the rates
	service.cacheRates(USD, rates)

	// Retrieve cached rates
	cached := service.getCachedRates(USD)
	if cached == nil {
		t.Fatal("Expected cached rates to be available")
	}

	if cached.Base != USD {
		t.Errorf("Expected cached base to be USD, got %s", cached.Base)
	}

	if len(cached.Rates) != 2 {
		t.Errorf("Expected 2 cached rates, got %d", len(cached.Rates))
	}

	if cached.Rates[EUR] != 0.85 {
		t.Errorf("Expected EUR rate to be 0.85, got %f", cached.Rates[EUR])
	}

	if cached.Rates[GBP] != 0.75 {
		t.Errorf("Expected GBP rate to be 0.75, got %f", cached.Rates[GBP])
	}
}

func TestCurrencyServiceImpl_FetchRatesFromAPI_Success(t *testing.T) {
	logger, _ := utils.NewSugardLogger()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ExchangeRatesResponse{
			Timestamp: time.Now().Unix(),
			Base:      USD,
			Rates: map[string]float64{
				EUR: 0.85,
				GBP: 0.75,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &CurrencyConfig{
		APIKey:             "test-key",
		BaseURL:            server.URL,
		CacheTTLHours:      24,
		BaseCurrency:       USD,
		HTTPTimeoutSeconds: 30,
	}

	service := NewServiceImpl(config, logger)

	rates, err := service.fetchRatesFromAPI(context.Background(), USD)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rates.Base != USD {
		t.Errorf("Expected base to be USD, got %s", rates.Base)
	}

	if len(rates.Rates) != 2 {
		t.Errorf("Expected 2 rates, got %d", len(rates.Rates))
	}

	if rates.Rates[EUR] != 0.85 {
		t.Errorf("Expected EUR rate to be 0.85, got %f", rates.Rates[EUR])
	}

	if rates.Rates[GBP] != 0.75 {
		t.Errorf("Expected GBP rate to be 0.75, got %f", rates.Rates[GBP])
	}
}

func TestCurrencyServiceImpl_FetchRatesFromAPI_NoAPIKey(t *testing.T) {
	logger := zap.NewNop().Sugar()
	config := &CurrencyConfig{
		APIKey:             "", // No API key
		BaseURL:            "https://test.com",
		CacheTTLHours:      24,
		BaseCurrency:       USD,
		HTTPTimeoutSeconds: 30,
	}

	service := NewServiceImpl(config, logger)

	_, err := service.fetchRatesFromAPI(context.Background(), USD)
	if err == nil {
		t.Error("Expected error when API key is not configured")
	}
}
