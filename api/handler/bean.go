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

package handler

import "strconv"

// Response represents a generic API response
type Response struct {
	Code   int         `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Errors []*ApiError `json:"errors,omitempty"`
}

// ApiError represents an error response
type ApiError struct {
	HttpStatusCode    int         `json:"-"`
	Code              string      `json:"code,omitempty"`
	InternalMessage   string      `json:"internalMessage,omitempty"`
	UserMessage       interface{} `json:"userMessage,omitempty"`
	UserDetailMessage string      `json:"userDetailMessage,omitempty"`
}

// NewApiError creates a new ApiError object
func NewApiError(httpStatusCode int, userMessage, internalMessage string) *ApiError {
	return &ApiError{
		HttpStatusCode:  httpStatusCode,
		Code:            strconv.Itoa(httpStatusCode),
		InternalMessage: internalMessage,
		UserMessage:     userMessage,
	}
}

// Error returns the internal message
func (e *ApiError) Error() string {
	return e.InternalMessage
}
