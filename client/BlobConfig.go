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

package util

import (
	"github.com/caarlos0/env"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"go.uber.org/zap"
)

type BlobConfigVariables struct {
	CloudConfigured           bool                         `env:"CLOUD_CONFIGURED" envDefault:"false"`
	BlobStorageType           blob_storage.BlobStorageType `env:"BLOB_STORAGE_TYPE"`
	S3AccessKey               string                       `env:"S3_ACCESS_KEY"`
	S3Passkey                 string                       `env:"S3_PASS_KEY"`
	S3EndpointUrl             string                       `env:"S3_END_POINT_URL"`
	S3IsInSecure              bool                         `env:"S3_IS_INSECURE"`
	S3BucketName              string                       `env:"S3_BUCKET_NAME"`
	S3Region                  string                       `env:"S3_REGION"`
	S3VersioningEnabled       bool                         `env:"S3_VERSIONING_ENABLED"`
	AzureEnabled              bool                         `env:"AZURE_ENABLED"`
	AzureAccountName          string                       `env:"AZURE_ACCOUNT_NAME"`
	AzureAccountKey           string                       `env:"AZURE_ACCOUNT_KEY"`
	AzureBlobContainerName    string                       `env:"AZURE_BLOB_CONTAINER_NAME"`
	GcpBucketName             string                       `env:"GCP_BUCKET_NAME"`
	GcpCredentialFileJsonData string                       `env:"GCP_CREDENTIAL_FILE_JSON_DATA"`

	// Feedback Storage Configuration
	FeedbackStorageType                blob_storage.BlobStorageType `env:"FEEDBACK_STORAGE_TYPE" envDefault:"S3"` // S3, GCP, AZURE
	FeedbackS3AccessKey                string                       `env:"FEEDBACK_S3_ACCESS_KEY"`
	FeedbackS3Passkey                  string                       `env:"FEEDBACK_S3_PASS_KEY"`
	FeedbackS3BucketName               string                       `env:"FEEDBACK_S3_BUCKET_NAME"`
	FeedbackS3Region                   string                       `env:"FEEDBACK_S3_REGION" envDefault:"us-east-1"`
	FeedbackGcpBucketName              string                       `env:"FEEDBACK_GCP_BUCKET_NAME"`
	FeedbackGcpCredentialFileJsonData  string                       `env:"FEEDBACK_GCP_CREDENTIAL_FILE_JSON_DATA"` // Can use same as Google Sheets service account
	FeedbackAzureEnabled               bool                         `env:"FEEDBACK_AZURE_ENABLED" envDefault:"false"`
	FeedbackAzureAccountName           string                       `env:"FEEDBACK_AZURE_ACCOUNT_NAME"`
	FeedbackAzureAccountKey            string                       `env:"FEEDBACK_AZURE_ACCOUNT_KEY"`
	FeedbackAzureBlobContainerName     string                       `env:"FEEDBACK_AZURE_BLOB_CONTAINER_NAME"`
}

func NewBlobConfig(logger *zap.SugaredLogger) (*BlobConfigVariables, error) {
	cfg := &BlobConfigVariables{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error on parsing module config", "err", err)
		return &BlobConfigVariables{}, err
	}
	return cfg, nil
}
