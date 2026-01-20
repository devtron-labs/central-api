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
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"
)

type DocsProxyHandler struct {
	logger *zap.SugaredLogger
	proxy  *httputil.ReverseProxy
}

func NewDocsProxyHandler(logger *zap.SugaredLogger) *DocsProxyHandler {
	// Get Python FastAPI server URL from environment or use default
	pythonServerURL := os.Getenv("DOCS_RAG_SERVER_URL")
	if pythonServerURL == "" {
		pythonServerURL = "http://localhost:8000"
	}

	targetURL, err := url.Parse(pythonServerURL)
	if err != nil {
		logger.Fatalw("Failed to parse DOCS_RAG_SERVER_URL", "url", pythonServerURL, "err", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Customize the director to strip the /docs prefix
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Strip /docs prefix from the path
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/docs")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		req.Host = targetURL.Host
		logger.Infow("Proxying request to Python FastAPI",
			"original_path", req.URL.Path,
			"target", targetURL.String())
	}

	// Add error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Errorw("Proxy error", "err", err, "path", r.URL.Path)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, `{"error": "Documentation service unavailable", "details": "%s"}`, err.Error())
	}

	logger.Infow("Docs proxy handler initialized", "target", pythonServerURL)

	return &DocsProxyHandler{
		logger: logger,
		proxy:  proxy,
	}
}

// ProxyRequest forwards the request to Python FastAPI server
func (h *DocsProxyHandler) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Infow("Proxying docs request", "method", r.Method, "path", r.URL.Path)
	h.proxy.ServeHTTP(w, r)
}
