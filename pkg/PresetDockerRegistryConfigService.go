package pkg

import (
	"github.com/devtron-labs/central-api/common"
)

type PresetDockerRegistryConfigService interface {
	GetConfig() *common.DockerRegistry
}

type PresetDockerRegistryConfigServiceImpl struct {
	config common.PresetDockerRegistryConfig
}

func (configService *PresetDockerRegistryConfigServiceImpl) GetConfig() *common.DockerRegistry {
	return &common.DockerRegistry{
		PluginId:           configService.config.PluginId,
		RegistryURL:        configService.config.RegistryURL,
		RegistryType:       configService.config.RegistryType,
		AWSAccessKeyId:     configService.config.AWSAccessKeyId,
		AWSSecretAccessKey: configService.config.AWSSecretAccessKey,
		AWSRegion:          configService.config.AWSRegion,
		Username:           configService.config.Username,
		Password:           configService.config.Password,
		IsDefault:          configService.config.IsDefault,
		Connection:         configService.config.Connection,
		Cert:               configService.config.Cert,
		Active:             configService.config.Active,
	}
}
