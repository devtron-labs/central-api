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

import (
	"encoding/json"
	"github.com/juju/errors"
	"log"
	"net/http"
)

// SetupCorsOriginHeader sets the CORS headers for the response.
//   - Use this function for all the public APIs.
func SetupCorsOriginHeader(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "text/html; charset=utf-8")
}

// WriteJsonResp writes the response in JSON format.
//   - Use this function for writing the response for all the APIs.
func WriteJsonResp(w http.ResponseWriter, err error, respBody interface{}, status int) {
	response := Response{}
	if err == nil {
		response.Result = respBody
	} else if apiErr, ok := err.(*ApiError); ok {
		response.Errors = []*ApiError{apiErr}
		if apiErr.HttpStatusCode != 0 {
			status = apiErr.HttpStatusCode
		}
	} else {
		apiErr := &ApiError{}
		apiErr.Code = "000" // 000=unknown
		apiErr.InternalMessage = err.Error()
		apiErr.InternalMessage = errors.Details(err)
		if respBody != nil {
			apiErr.UserMessage = respBody
		} else {
			apiErr.UserMessage = err.Error()
		}
		response.Errors = []*ApiError{apiErr}
	}
	response.Code = status
	response.Status = http.StatusText(status)
	b, err := json.Marshal(response)
	if err != nil {
		log.Println("error in marshaling err object", "err", err)
		status = 500

		response := Response{}
		apiErr := &ApiError{}
		apiErr.Code = "0000" // 000=unknown
		apiErr.InternalMessage = errors.Details(err)
		apiErr.UserMessage = "response marshaling error"
		response.Errors = []*ApiError{apiErr}
		b, err = json.Marshal(response)
		if err != nil {
			b = []byte("response marshaling error")
			log.Println("Unexpected error in apiError", "err", err)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(b)
	if err != nil {
		log.Println("error in writing response", "err", err)
	}
}
