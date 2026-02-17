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
	"math"
)

// ValidateCurrencyCode checks if a currency code is valid (3-letter ISO format)
func ValidateCurrencyCode(code string) bool {
	if len(code) != 3 {
		return false
	}

	// Check if all characters are uppercase letters
	for _, char := range code {
		if char < 'A' || char > 'Z' {
			return false
		}
	}

	return true
}

// FormatCurrency formats a currency amount to a reasonable number of decimal places
func FormatCurrency(amount float64) float64 {
	// Round to 6 decimal places for precision
	return math.Round(amount*1000000) / 1000000
}
