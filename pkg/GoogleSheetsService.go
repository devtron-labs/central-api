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
	"context"
	"fmt"
	"strings"
	util "github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/common"
	"go.uber.org/zap"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService interface {
	AddRow(data *common.FeedbackData) error
	AddRowToSheet(sheetName string, data *common.FeedbackData) error
	IsConfigured() bool
}

type GoogleSheetsServiceImpl struct {
	logger             *zap.SugaredLogger
	googleSheetsClient *util.GoogleSheetsClient
}

func NewGoogleSheetsServiceImpl(logger *zap.SugaredLogger, googleSheetsClient *util.GoogleSheetsClient) *GoogleSheetsServiceImpl {
	return &GoogleSheetsServiceImpl{
		logger:             logger,
		googleSheetsClient: googleSheetsClient,
	}
}

// AddRow adds a new row to the default sheet in the configured spreadsheet
func (impl *GoogleSheetsServiceImpl) AddRow(data *common.FeedbackData) error {
	return impl.AddRowToSheet("Sheet1", data)
}

// IsConfigured returns true if Google Sheets is properly configured
func (impl *GoogleSheetsServiceImpl) IsConfigured() bool {
	return impl.googleSheetsClient.IsConfigured()
}

// AddRowToSheet adds a new row to a specific sheet in the configured spreadsheet
func (impl *GoogleSheetsServiceImpl) AddRowToSheet(sheetName string, data *common.FeedbackData) error {
	if !impl.googleSheetsClient.IsConfigured() {
		impl.logger.Warn("Google Sheets client not configured, skipping row addition")
		return fmt.Errorf("google Sheets not configured")
	}

	// Prepare the row data (starting from column A)
	// Columns: A=UCID, B=ThreadName, C=UserEmail, D=Reasons, E=AdditionalDetails, F=SubmittedAt, G=FullConversationURL
	// Note: ConversationText is NOT included in the sheet (it's stored in S3/GCP)
	reasonsStr := ""
	if len(data.Reasons) > 0 {
		reasonsStr = strings.Join(data.Reasons, ", ")
	}

	values := []interface{}{
		data.UCID,
		data.ThreadName,
		data.UserEmail,
		reasonsStr,
		data.AdditionalDetails,
		data.SubmittedAt.Format("2006-01-02 15:04:05"),
		data.FullConversationURL,
	}

	// Create the value range
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	// Append the row to the sheet (A:G - 7 columns)
	spreadsheetId := impl.googleSheetsClient.Config.SpreadsheetID
	rangeToAppend := fmt.Sprintf("%s!A:G", sheetName)

	_, err := impl.googleSheetsClient.SheetsService.Spreadsheets.Values.Append(
		spreadsheetId,
		rangeToAppend,
		valueRange,
	).ValueInputOption("RAW").Context(context.Background()).Do()

	if err != nil {
		impl.logger.Errorw("error appending row to Google Sheets",
			"err", err,
			"spreadsheetId", spreadsheetId,
			"sheetName", sheetName,
		)
		return err
	}

	impl.logger.Infow("successfully added row to Google Sheets",
		"spreadsheetId", spreadsheetId,
		"sheetName", sheetName,
		"ucid", data.UCID,
	)

	return nil
}
