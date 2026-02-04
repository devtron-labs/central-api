/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package util

import (
	"context"
	"encoding/json"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GoogleCloudConfig holds configuration for Google Cloud services (Sheets, Storage, etc.)
type GoogleCloudConfig struct {
	ServiceAccountJSON string `env:"FEEDBACK_GCP_CREDENTIAL_FILE_JSON_DATA" envDefault:""`
	SpreadsheetID      string `env:"GOOGLE_SPREADSHEET_ID" envDefault:""`
}

// GoogleSheetsClient provides access to Google Sheets API and service account credentials
// The same service account can be used for Google Cloud Storage
type GoogleSheetsClient struct {
	SheetsService *sheets.Service
	Config        *GoogleCloudConfig
}

/* #nosec */
func NewGoogleSheetsClient(logger *zap.SugaredLogger, blobConfig *BlobConfigVariables) (*GoogleSheetsClient, error) {
	cfg := &GoogleCloudConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error parsing google cloud config", "err", err)
		return &GoogleSheetsClient{}, err
	}

	// Fallback: If GOOGLE_SERVICE_ACCOUNT_JSON is not set, try using FEEDBACK_GCP_CREDENTIAL_FILE_JSON_DATA
	serviceAccountJSON := cfg.ServiceAccountJSON
	if serviceAccountJSON == "" && blobConfig != nil && blobConfig.FeedbackGcpCredentialFileJsonData != "" {
		serviceAccountJSON = blobConfig.FeedbackGcpCredentialFileJsonData
		logger.Info("Using FEEDBACK_GCP_CREDENTIAL_FILE_JSON_DATA for Google Sheets authentication")
	}

	// If service account JSON is not provided, return empty client
	if serviceAccountJSON == "" {
		logger.Warn("Google service account JSON not provided (neither GOOGLE_SERVICE_ACCOUNT_JSON nor FEEDBACK_GCP_CREDENTIAL_FILE_JSON_DATA), Google Sheets client will not be initialized")
		return &GoogleSheetsClient{
			SheetsService: nil,
			Config:        cfg,
		}, nil
	}

	// Store the actual service account JSON being used
	cfg.ServiceAccountJSON = serviceAccountJSON

	ctx := context.Background()

	// Parse the service account JSON
	credentials, err := google.CredentialsFromJSON(ctx, []byte(serviceAccountJSON), sheets.SpreadsheetsScope)
	if err != nil {
		logger.Errorw("error creating credentials from service account JSON", "err", err)
		return nil, err
	}

	// Create the sheets service
	sheetsService, err := sheets.NewService(ctx, option.WithCredentials(credentials))
	if err != nil {
		logger.Errorw("error creating sheets service", "err", err)
		return nil, err
	}

	logger.Info("Google Sheets client initialized successfully")

	return &GoogleSheetsClient{
		SheetsService: sheetsService,
		Config:        cfg,
	}, nil
}

// IsConfigured returns true if the Google Sheets client is properly configured
func (c *GoogleSheetsClient) IsConfigured() bool {
	if c == nil {
		return false
	}
	return c.SheetsService != nil && c.Config != nil && c.Config.SpreadsheetID != ""
}

// GetSpreadsheetID returns the configured spreadsheet ID
func (c *GoogleSheetsClient) GetSpreadsheetID() string {
	if c == nil || c.Config == nil {
		return ""
	}
	return c.Config.SpreadsheetID
}

// GetServiceAccountJSON returns the service account JSON for use with other Google Cloud services
func (c *GoogleSheetsClient) GetServiceAccountJSON() string {
	if c == nil || c.Config == nil {
		return ""
	}
	return c.Config.ServiceAccountJSON
}

// MarshalServiceAccountJSON is a helper to convert service account key to JSON string
func MarshalServiceAccountJSON(serviceAccountKey map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(serviceAccountKey)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
