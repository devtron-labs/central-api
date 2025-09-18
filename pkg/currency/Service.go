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
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Service defines the interface for currency operations
type Service interface {

	// GetExchangeRates returns the latest exchange rates for the base currency
	GetExchangeRates(ctx context.Context, base string) (*ExchangeRatesResponse, error)

	// RefreshRates manually refreshes the cached exchange rates
	RefreshRatesForDefaultBase(ctx context.Context) error
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	config     *CurrencyConfig
	logger     *zap.SugaredLogger
	httpClient *http.Client
	cache      *sync.Map // map[string]*CachedRates
	cacheMutex sync.RWMutex
}

// NewServiceImpl creates a new instance of ServiceImpl
func NewServiceImpl(config *CurrencyConfig, logger *zap.SugaredLogger) *ServiceImpl {
	httpClient := &http.Client{
		Timeout: time.Duration(config.HTTPTimeoutSeconds) * time.Second,
	}

	service := &ServiceImpl{
		config:     config,
		logger:     logger,
		httpClient: httpClient,
		cache:      &sync.Map{},
	}

	// Initialize cache with default base currency
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorw("GO_ROUTINE_PANIC_LOG: go-routine recovered from panic", "err", r, "stack", string(debug.Stack()))
			}
		}()
		ctx := context.Background()
		if err := service.RefreshRatesForDefaultBase(ctx); err != nil {
			logger.Errorw("failed to initialize currency cache", "err", err)
		}
	}()

	return service
}

// GetExchangeRates returns the latest exchange rates for the specified base currency
func (s *ServiceImpl) GetExchangeRates(ctx context.Context, base string) (*ExchangeRatesResponse, error) {
	if base == "" {
		base = s.config.BaseCurrency
	}

	// Try to get from cache first
	if cachedRates := s.getCachedRates(base); cachedRates != nil {
		s.logger.Debugw("returning cached exchange rates", "base", base, "cacheExpiry", cachedRates.ExpiresAt)
		return &ExchangeRatesResponse{
			Timestamp: cachedRates.Timestamp.Unix(),
			Base:      cachedRates.Base,
			Rates:     cachedRates.Rates,
		}, nil
	}

	// Fetch from API if not in cache or expired
	rates, err := s.fetchRatesFromAPI(ctx, base)
	if err != nil {
		s.logger.Errorw("failed to fetch rates from API", "base", base, "err", err)

		// Try to return stale cache data if available
		if staleRates := s.getStaleRates(base); staleRates != nil {
			s.logger.Warnw("returning stale cached rates due to API failure", "base", base)
			return &ExchangeRatesResponse{
				Timestamp: staleRates.Timestamp.Unix(),
				Base:      staleRates.Base,
				Rates:     staleRates.Rates,
			}, nil
		}

		return nil, fmt.Errorf("failed to fetch exchange rates and no cached data available: %w", err)
	}

	// Cache the new rates
	s.cacheRates(base, rates)

	return rates, nil
}

// RefreshRatesForDefaultBase manually refreshes the cached exchange rates for the default base currency
func (s *ServiceImpl) RefreshRatesForDefaultBase(ctx context.Context) error {
	s.logger.Info("refreshing currency exchange rates")

	// Refresh rates for the default base currency
	rates, err := s.fetchRatesFromAPI(ctx, s.config.BaseCurrency)
	if err != nil {
		s.logger.Errorw("failed to refresh rates", "base", s.config.BaseCurrency, "err", err)
		return err
	}
	// Cache the new rates
	s.cacheRates(s.config.BaseCurrency, rates)
	s.logger.Info("currency exchange rates refreshed successfully")
	return nil
}

// fetchRatesFromAPI fetches exchange rates from the OpenExchangeRates API
func (s *ServiceImpl) fetchRatesFromAPI(ctx context.Context, base string) (*ExchangeRatesResponse, error) {
	if s.config.APIKey == "" {
		return nil, fmt.Errorf("OpenExchangeRates API key not configured")
	}
	baseURl, err := url.Parse(s.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	exchangeRateUrl := baseURl.JoinPath("latest.json")
	exchangeRateUrl.Query().Set("app_id", s.config.APIKey)
	if base != USD {
		exchangeRateUrl.Query().Set("base", base)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", exchangeRateUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Devtron-Central-API/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var rates ExchangeRatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	s.logger.Debugw("fetched rates from API",
		"base", rates.Base,
		"timestamp", rates.Timestamp,
		"rateCount", len(rates.Rates))

	return &rates, nil
}

// getCachedRates retrieves valid (non-expired) cached rates
func (s *ServiceImpl) getCachedRates(base string) *CachedRates {
	value, exists := s.cache.Load(base)
	if !exists {
		return nil
	}

	cached, ok := value.(*CachedRates)
	if !ok {
		return nil
	}

	// Check if cache is still valid
	if time.Now().After(cached.ExpiresAt) {
		s.logger.Debugw("cached rates expired", "base", base, "expiredAt", cached.ExpiresAt)
		return nil
	}

	return cached
}

// getStaleRates retrieves cached rates even if expired (for fallback purposes)
func (s *ServiceImpl) getStaleRates(base string) *CachedRates {
	value, exists := s.cache.Load(base)
	if !exists {
		return nil
	}

	cached, ok := value.(*CachedRates)
	if !ok {
		return nil
	}

	return cached
}

// cacheRates stores exchange rates in the cache with TTL
func (s *ServiceImpl) cacheRates(base string, rates *ExchangeRatesResponse) {
	cached := &CachedRates{
		Rates:     rates.Rates,
		Base:      rates.Base,
		Timestamp: time.Unix(rates.Timestamp, 0),
		ExpiresAt: time.Now().Add(time.Duration(s.config.CacheTTLHours) * time.Hour),
	}

	s.cache.Store(base, cached)

	s.logger.Debugw("cached exchange rates",
		"base", base,
		"expiresAt", cached.ExpiresAt,
		"rateCount", len(cached.Rates))
}
