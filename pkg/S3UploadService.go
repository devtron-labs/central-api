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

package pkg

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	util "github.com/devtron-labs/central-api/client"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type S3UploadService interface {
	UploadFile(content string, fileName string) (string, error)
	IsConfigured() bool
}

type S3UploadServiceImpl struct {
	logger             *zap.SugaredLogger
	blobConfig         *util.BlobConfigVariables
	googleSheetsClient *util.GoogleSheetsClient
	s3Client           *s3.S3
	gcpStorageClient   *storage.Client
	storageType        blob_storage.BlobStorageType
}

func NewS3UploadServiceImpl(logger *zap.SugaredLogger, blobConfig *util.BlobConfigVariables, googleSheetsClient *util.GoogleSheetsClient) (*S3UploadServiceImpl, error) {
	var s3Client *s3.S3
	var gcpStorageClient *storage.Client
	storageType := blobConfig.FeedbackStorageType

	logger.Infow("initializing feedback storage service", "storageType", storageType)

	switch storageType {
	case blob_storage.BLOB_STORAGE_S3:
		// Initialize S3 client
		if blobConfig.FeedbackS3BucketName != "" && blobConfig.FeedbackS3AccessKey != "" {
			awsConfig := &aws.Config{
				Region: aws.String(blobConfig.FeedbackS3Region),
				Credentials: credentials.NewStaticCredentials(
					blobConfig.FeedbackS3AccessKey,
					blobConfig.FeedbackS3Passkey,
					"",
				),
			}

			sess, err := session.NewSession(awsConfig)
			if err != nil {
				logger.Errorw("error creating AWS session for feedback storage", "err", err)
				return nil, err
			}

			s3Client = s3.New(sess)
			logger.Info("Feedback S3 client initialized successfully")
		} else {
			logger.Warn("Feedback S3 not configured properly")
		}

	case blob_storage.BLOB_STORAGE_GCP:
		// Initialize GCP Storage client
		ctx := context.Background()
		var err error

		// Try to use feedback-specific GCP credentials first, fall back to Google Sheets service account
		credentialJSON := blobConfig.FeedbackGcpCredentialFileJsonData
		if credentialJSON == "" && googleSheetsClient != nil {
			credentialJSON = googleSheetsClient.GetServiceAccountJSON()
			logger.Info("Using Google Sheets service account for GCP Storage")
		}

		if credentialJSON != "" && blobConfig.FeedbackGcpBucketName != "" {
			gcpStorageClient, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(credentialJSON)))
			if err != nil {
				logger.Errorw("error creating GCP storage client", "err", err)
				return nil, err
			}
			logger.Info("Feedback GCP Storage client initialized successfully")
		} else {
			logger.Warn("Feedback GCP Storage not configured properly")
		}

	case blob_storage.BLOB_STORAGE_AZURE:
		// Azure support can be added here if needed
		logger.Warn("Azure blob storage for feedback is configured but not yet implemented")

	default:
		logger.Warnw("unknown or unsupported feedback storage type", "storageType", storageType)
	}

	return &S3UploadServiceImpl{
		logger:             logger,
		blobConfig:         blobConfig,
		googleSheetsClient: googleSheetsClient,
		s3Client:           s3Client,
		gcpStorageClient:   gcpStorageClient,
		storageType:        storageType,
	}, nil
}

// IsConfigured returns true if any storage client is configured
func (impl *S3UploadServiceImpl) IsConfigured() bool {
	switch impl.storageType {
	case blob_storage.BLOB_STORAGE_S3:
		return impl.s3Client != nil
	case blob_storage.BLOB_STORAGE_GCP:
		return impl.gcpStorageClient != nil
	case blob_storage.BLOB_STORAGE_AZURE:
		return false // Not implemented yet
	default:
		return false
	}
}

// UploadFile uploads content to the configured storage provider and returns the URL
func (impl *S3UploadServiceImpl) UploadFile(content string, fileName string) (string, error) {
	switch impl.storageType {
	case blob_storage.BLOB_STORAGE_S3:
		return impl.uploadToS3(content, fileName)
	case blob_storage.BLOB_STORAGE_GCP:
		return impl.uploadToGCP(content, fileName)
	case blob_storage.BLOB_STORAGE_AZURE:
		return "", fmt.Errorf("Azure blob storage not yet implemented")
	default:
		return "", fmt.Errorf("unsupported storage type: %s", impl.storageType)
	}
}

// uploadToS3 uploads content to AWS S3
func (impl *S3UploadServiceImpl) uploadToS3(content string, fileName string) (string, error) {
	if impl.s3Client == nil {
		impl.logger.Warn("S3 client not configured, skipping upload")
		return "", fmt.Errorf("S3 not configured")
	}

	// Upload to S3
	_, err := impl.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(impl.blobConfig.FeedbackS3BucketName),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader([]byte(content)),
		ContentType: aws.String("text/plain"),
	})

	if err != nil {
		impl.logger.Errorw("error uploading file to S3", "err", err, "fileName", fileName)
		return "", err
	}

	// Construct S3 URL
	s3URL := impl.constructS3URL(fileName)

	impl.logger.Infow("successfully uploaded file to S3", "fileName", fileName, "url", s3URL)
	return s3URL, nil
}

// uploadToGCP uploads content to Google Cloud Storage
func (impl *S3UploadServiceImpl) uploadToGCP(content string, fileName string) (string, error) {
	if impl.gcpStorageClient == nil {
		impl.logger.Warn("GCP Storage client not configured, skipping upload")
		return "", fmt.Errorf("GCP Storage not configured")
	}

	ctx := context.Background()
	bucket := impl.gcpStorageClient.Bucket(impl.blobConfig.FeedbackGcpBucketName)
	obj := bucket.Object(fileName)

	// Create a writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = "text/plain"

	// Write content
	_, err := writer.Write([]byte(content))
	if err != nil {
		impl.logger.Errorw("error writing to GCP Storage", "err", err, "fileName", fileName)
		writer.Close()
		return "", err
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		impl.logger.Errorw("error closing GCP Storage writer", "err", err, "fileName", fileName)
		return "", err
	}

	// Construct GCP Storage URL
	gcpURL := impl.constructGCPURL(fileName)

	impl.logger.Infow("successfully uploaded file to GCP Storage", "fileName", fileName, "url", gcpURL)
	return gcpURL, nil
}

// constructS3URL constructs the S3 URL for the uploaded file
func (impl *S3UploadServiceImpl) constructS3URL(fileName string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		impl.blobConfig.FeedbackS3BucketName,
		impl.blobConfig.FeedbackS3Region,
		fileName)
}

// constructGCPURL constructs the GCP Storage URL for the uploaded file
func (impl *S3UploadServiceImpl) constructGCPURL(fileName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s",
		impl.blobConfig.FeedbackGcpBucketName,
		fileName)
}
