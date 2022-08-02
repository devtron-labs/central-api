package common

import (
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type PresetDockerRegistryConfig struct {
	PluginId           string `env:"DOCKER_REGISTRY_PLUGIN_ID" envDefault:"cd.go.artifact.docker.registry"`
	RegistryURL        string `env:"DOCKER_REGISTRY_URL" envDefault:"ttl.sh"`
	RegistryType       string `env:"DOCKER_REGISTRY_TYPE" envDefault:"other"`
	AWSAccessKeyId     string `env:"DOCKER_REGISTRY_AWS_ACCESS_KEY" envDefault:""`
	AWSSecretAccessKey string `env:"DOCKER_REGISTRY_AWS_SECRET_KEY" envDefault:""`
	AWSRegion          string `env:"DOCKER_REGISTRY_AWS_REGION" envDefault:""`
	Username           string `env:"DOCKER_REGISTRY_USERNAME" envDefault:"a"`
	Password           string `env:"DOCKER_REGISTRY_PASSWORD" envDefault:"a"`
	IsDefault          bool   `env:"DOCKER_REGISTRY_IS_DEFAULT" envDefault:"false"`
	Connection         string `env:"DOCKER_REGISTRY_CONNECTION" envDefault:"secure"`
	Cert               string `env:"DOCKER_REGISTRY_CERT" envDefault:""`
	Active             bool   `env:"DOCKER_REGISTRY_ACTIVE" envDefault:"true"`
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
