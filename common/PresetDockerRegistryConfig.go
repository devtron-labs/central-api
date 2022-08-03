package common

import (
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type PresetDockerRegistryConfig struct {
	PluginId               string `env:"PRESET_REGISTRY_PLUGIN_ID" envDefault:"cd.go.artifact.docker.registry"`
	RegistryURL            string `env:"PRESET_REGISTRY_URL" envDefault:"ttl.sh"`
	RegistryType           string `env:"PRESET_REGISTRY_TYPE" envDefault:"other"`
	AWSAccessKeyId         string `env:"PRESET_REGISTRY_AWS_ACCESS_KEY" envDefault:""`
	AWSSecretAccessKey     string `env:"PRESET_REGISTRY_AWS_SECRET_KEY" envDefault:""`
	AWSRegion              string `env:"PRESET_REGISTRY_AWS_REGION" envDefault:""`
	Username               string `env:"PRESET_REGISTRY_USERNAME" envDefault:"a"`
	Password               string `env:"PRESET_REGISTRY_PASSWORD" envDefault:"a"`
	IsDefault              bool   `env:"PRESET_REGISTRY_IS_DEFAULT" envDefault:"false"`
	Connection             string `env:"PRESET_REGISTRY_CONNECTION" envDefault:"secure"`
	Cert                   string `env:"PRESET_REGISTRY_CERT" envDefault:""`
	Active                 bool   `env:"PRESET_REGISTRY_ACTIVE" envDefault:"true"`
	ExpiryTime             int    `env:"PRESET_REGISTRY_IMAGE_EXPIRY_TIME_SECs" envDefault:"86400"`
	PresetRegistryRepoName string `env:"PRESET_REGISTRY_REPO_NAME" envDefault:"devtron-preset-registry-repo"`
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
