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
