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
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"github.com/devtron-labs/central-api/common"
	"go.uber.org/zap"
	"io"
	"strings"
	"time"
)

type FeedbackService interface {
	SubmitFeedback(feedbackData *common.FeedbackData) error
}

type FeedbackServiceImpl struct {
	logger              *zap.SugaredLogger
	s3UploadService     S3UploadService
	googleSheetsService GoogleSheetsService
}

func NewFeedbackServiceImpl(
	logger *zap.SugaredLogger,
	s3UploadService S3UploadService,
	googleSheetsService GoogleSheetsService,
) *FeedbackServiceImpl {
	return &FeedbackServiceImpl{
		logger:              logger,
		s3UploadService:     s3UploadService,
		googleSheetsService: googleSheetsService,
	}
}

// SubmitFeedback orchestrates the feedback submission process:
// 1. Decompresses conversation text if compressed
// 2. Uploads conversation content to S3
// 3. Adds feedback data with S3 URL to Google Sheets
func (impl *FeedbackServiceImpl) SubmitFeedback(feedbackData *common.FeedbackData) error {
	impl.logger.Infow("submitting feedback", "ucid", feedbackData.UCID, "threadName", feedbackData.ThreadName, "isCompressed", feedbackData.IsCompressed)

	// Step 1: Upload conversation content to S3
	var s3URL string
	var err error

	if impl.s3UploadService.IsConfigured() && feedbackData.ConversationText != "" {
		// Decompress conversation text if it's compressed
		conversationContent := feedbackData.ConversationText
		if feedbackData.IsCompressed {
			conversationContent, err = decompressConversationText(feedbackData.ConversationText)
			if err != nil {
				impl.logger.Errorw("error decompressing conversation text", "err", err, "ucid", feedbackData.UCID)
				return fmt.Errorf("failed to decompress conversation text: %w", err)
			}
			impl.logger.Infow("successfully decompressed conversation text", "ucid", feedbackData.UCID,
				"originalSize", len(feedbackData.ConversationText), "decompressedSize", len(conversationContent))
		}

		// Sanitize conversation name and create filename
		sanitizedName := sanitizeFileName(feedbackData.ThreadName)
		fileName := fmt.Sprintf("%s-%s.txt", sanitizedName, feedbackData.UCID)

		s3URL, err = impl.s3UploadService.UploadFile(conversationContent, fileName)
		if err != nil {
			impl.logger.Errorw("error uploading feedback to S3", "err", err, "ucid", feedbackData.UCID)
			// Continue even if S3 upload fails - we still want to save to Google Sheets
			s3URL = ""
		} else {
			impl.logger.Infow("successfully uploaded feedback to S3", "ucid", feedbackData.UCID, "url", s3URL)
		}
	} else {
		impl.logger.Warn("S3 upload service not configured or conversation text is empty, skipping S3 upload")
	}

	// Update feedback data with S3 URL
	feedbackData.FullConversationURL = s3URL

	// Set submitted time if not already set
	if feedbackData.SubmittedAt.IsZero() {
		feedbackData.SubmittedAt = time.Now().UTC()
	}

	// Step 2: Add to Google Sheets (optional - log error but don't fail the request)
	if impl.googleSheetsService.IsConfigured() {
		err = impl.googleSheetsService.AddRow(feedbackData)
		if err != nil {
			impl.logger.Errorw("error adding feedback to Google Sheets", "err", err, "ucid", feedbackData.UCID)
			// Don't return error - Google Sheets is optional
			// The S3 upload already succeeded, so we consider the feedback submitted
		} else {
			impl.logger.Infow("successfully added feedback to Google Sheets", "ucid", feedbackData.UCID)
		}
	} else {
		impl.logger.Warn("Google Sheets not configured, skipping sheet update", "ucid", feedbackData.UCID)
	}

	impl.logger.Infow("successfully submitted feedback", "ucid", feedbackData.UCID, "s3URL", s3URL)
	return nil
}

// decompressConversationText decompresses zlib-compressed conversation text
// The input is expected to be a base64-encoded zlib-compressed string from Python:
// Python: base64.b64encode(zlib.compress(text.encode('utf-8'))).decode('utf-8')
func decompressConversationText(compressedText string) (string, error) {
	// Validate input
	if compressedText == "" {
		return "", fmt.Errorf("compressed text is empty")
	}

	// Step 1: Decode from base64
	compressedBytes, err := base64.StdEncoding.DecodeString(compressedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 (length: %d): %w", len(compressedText), err)
	}

	// Validate we have data after base64 decode
	if len(compressedBytes) == 0 {
		return "", fmt.Errorf("base64 decoded to empty bytes")
	}

	// Step 2: Decompress using zlib
	reader, err := zlib.NewReader(bytes.NewReader(compressedBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create zlib reader (compressed size: %d bytes): %w", len(compressedBytes), err)
	}
	defer reader.Close()

	// Step 3: Read decompressed data
	var decompressed bytes.Buffer
	_, err = io.Copy(&decompressed, reader)
	if err != nil {
		return "", fmt.Errorf("failed to decompress data: %w", err)
	}

	return decompressed.String(), nil
}

// sanitizeFileName removes or replaces characters that are not file-system friendly
func sanitizeFileName(name string) string {
	// Replace spaces and special characters with hyphens
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	sanitized := replacer.Replace(name)

	// Remove consecutive hyphens
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}

	// Trim hyphens from start and end
	sanitized = strings.Trim(sanitized, "-")

	return sanitized
}
