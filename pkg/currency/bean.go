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

import "time"

// currency codes constants
const (
	USD = "USD" // US Dollar
	EUR = "EUR" // Euro
	GBP = "GBP" // British Pound Sterling
	JPY = "JPY" // Japanese Yen
	AUD = "AUD" // Australian Dollar
	CAD = "CAD" // Canadian Dollar
	CHF = "CHF" // Swiss Franc
	CNY = "CNY" // Chinese Yuan
	SEK = "SEK" // Swedish Krona
	NZD = "NZD" // New Zealand Dollar
	MXN = "MXN" // Mexican Peso
	SGD = "SGD" // Singapore Dollar
	HKD = "HKD" // Hong Kong Dollar
	NOK = "NOK" // Norwegian Krone
	TRY = "TRY" // Turkish Lira
	RUB = "RUB" // Russian Ruble
	INR = "INR" // Indian Rupee
	BRL = "BRL" // Brazilian Real
	ZAR = "ZAR" // South African Rand
	KRW = "KRW" // South Korean Won
)

// GetCommonCurrencies returns a list of commonly used currency codes
func GetCommonCurrencies() []string {
	return []string{
		USD, EUR, GBP, JPY, AUD, CAD, CHF, CNY, SEK, NZD,
		MXN, SGD, HKD, NOK, TRY, RUB, INR, BRL, ZAR, KRW,
	}
}

// ExchangeRatesResponse represents the response from OpenExchangeRates API
type ExchangeRatesResponse struct {
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
}

// CachedRates represents cached exchange rates with expiration
type CachedRates struct {
	Rates     map[string]float64 `json:"rates"`
	Base      string             `json:"base"`
	Timestamp time.Time          `json:"timestamp"`
	ExpiresAt time.Time          `json:"expires_at"`
}
