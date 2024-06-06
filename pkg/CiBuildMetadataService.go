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

package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/central-api/common"
	"go.uber.org/zap"
	"os"
)

type CiBuildMetadataService interface {
	GetDockerfileTemplateMetadata() *common.DockerfileTemplateMetadata
	GetBuildpackMetadata() *common.BuildPackMetadata
}

type CiBuildMetadataServiceImpl struct {
	Logger                     *zap.SugaredLogger
	BuildPackMetadata          *common.BuildPackMetadata
	DockerfileTemplateMetadata *common.DockerfileTemplateMetadata
}

func NewCiBuildMetadataServiceImpl(logger *zap.SugaredLogger) *CiBuildMetadataServiceImpl {
	buildpackMetadata := setupBuildpackMetadata()
	templateMetadata := setupDockerfileTemplateMetadata()
	metadataServiceImpl := &CiBuildMetadataServiceImpl{
		Logger:                     logger,
		BuildPackMetadata:          buildpackMetadata,
		DockerfileTemplateMetadata: templateMetadata,
	}
	return metadataServiceImpl
}

func setupDockerfileTemplateMetadata() *common.DockerfileTemplateMetadata {

	dockerfileTemplateData, err := os.ReadFile("/DockerfileTemplateData.json")
	if err != nil {
		fmt.Println("error occurred while reading file DockerfileTemplateData.json", "error", err)
		return nil
	}
	dockerfileTemplateMetadata := &common.DockerfileTemplateMetadata{}
	err = json.Unmarshal(dockerfileTemplateData, dockerfileTemplateMetadata)
	if err != nil {
		fmt.Println("error occurred while unmarshalling json", "data", string(dockerfileTemplateData), "err", err)
		return nil
	}
	return dockerfileTemplateMetadata
}

func setupBuildpackMetadata() *common.BuildPackMetadata {

	buildpackMetadataBytes, err := os.ReadFile("/BuildpackMetadata.json")
	if err != nil {
		fmt.Println("error occurred while reading file DockerfileTemplateData.json", "error", err)
		return nil
	}
	buildpackMetadata := &common.BuildPackMetadata{}
	err = json.Unmarshal(buildpackMetadataBytes, buildpackMetadata)
	if err != nil {
		fmt.Println("error occurred while unmarshalling buildpack json", "data", string(buildpackMetadataBytes), "err", err)
		return nil
	}
	return buildpackMetadata
}

func (impl CiBuildMetadataServiceImpl) GetDockerfileTemplateMetadata() *common.DockerfileTemplateMetadata {
	return impl.DockerfileTemplateMetadata
}

func (impl CiBuildMetadataServiceImpl) GetBuildpackMetadata() *common.BuildPackMetadata {
	return impl.BuildPackMetadata
}
