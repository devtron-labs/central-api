package common

import (
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type PresetDockerRegistryConfig struct {
	PluginId           string `env:"DOCKER_REGISTRY_PLUGIN_ID" envDefault:""`
	RegistryURL        string `env:"DOCKER_REGISTRY_URL" envDefault:""`
	RegistryType       string `env:"DOCKER_REGISTRY_TYPE" envDefault:""`
	AWSAccessKeyId     string `env:"DOCKER_REGISTRY_AWS_ACCESS_KEY" envDefault:""`
	AWSSecretAccessKey string `env:"DOCKER_REGISTRY_AWS_SECRET_KEY" envDefault:""`
	AWSRegion          string `env:"DOCKER_REGISTRY_AWS_REGION" envDefault:""`
	Username           string `env:"DOCKER_REGISTRY_USERNAME" envDefault:""`
	Password           string `env:"DOCKER_REGISTRY_PASSWORD" envDefault:""`
	IsDefault          bool   `env:"DOCKER_REGISTRY_IS_DEFAULT" envDefault:""`
	Connection         string `env:"DOCKER_REGISTRY_CONNECTION" envDefault:""`
	Cert               string `env:"DOCKER_REGISTRY_CERT" envDefault:""`
	Active             bool   `env:"DOCKER_REGISTRY_ACTIVE" envDefault:""`
}

func NewPresetDockerRegistryConfig(logger *zap.SugaredLogger) (*PresetDockerRegistryConfig, error) {

	cfg := &PresetDockerRegistryConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error on parsing docker preset container registry config", "err", err)
		return &PresetDockerRegistryConfig{}, err
	}
	return cfg, nil
}
