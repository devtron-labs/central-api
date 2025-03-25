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

package releaseNote

import (
	"github.com/devtron-labs/central-api/common"
	"github.com/devtron-labs/central-api/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

// Deprecated: db is not in use anymore, we maintain cache and fetch data from cache
type ReleaseNote struct {
	tableName   struct{}          `sql:"release_notes"`
	Id          int               `sql:"id,pk"`
	ReleaseNote []*common.Release `sql:"release_note, notnull"`
	IsActive    bool              `sql:"is_active, notnull"`
	CreatedOn   time.Time         `sql:"created_on,type:timestamptz,notnull"`
	UpdatedOn   time.Time         `sql:"updated_on,type:timestamptz"`
}

type ReleaseNoteRepository interface {
	GetConnection() *pg.DB
	Save(releaseNote *ReleaseNote, tx *pg.Tx) error
	Update(releaseNote *ReleaseNote, tx *pg.Tx) error
	FindActive() (*ReleaseNote, error)
}

type ReleaseNoteRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewReleaseNoteRepositoryImpl(logger *zap.SugaredLogger) (*ReleaseNoteRepositoryImpl, error) {
	dbConnection, err := sql.NewDbConnection(logger)
	if err != nil {
		return nil, err
	}
	return &ReleaseNoteRepositoryImpl{dbConnection: dbConnection}, nil
}

func (impl ReleaseNoteRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl ReleaseNoteRepositoryImpl) Save(releaseNote *ReleaseNote, tx *pg.Tx) error {
	return tx.Insert(releaseNote)
}

func (impl ReleaseNoteRepositoryImpl) Update(releaseNote *ReleaseNote, tx *pg.Tx) error {
	return tx.Update(releaseNote)
}

func (impl ReleaseNoteRepositoryImpl) FindActive() (*ReleaseNote, error) {
	releaseNote := &ReleaseNote{}
	err := impl.dbConnection.Model(releaseNote).
		Where("is_active = ?", true).
		Select()
	return releaseNote, err
}
