package config

import (
	_ "embed"
	"errors"

	"github.com/pupload/pupload/internal/controller/flows/repo"
	"github.com/pupload/pupload/internal/controller/projects"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
	"github.com/spf13/viper"
)

type ControllerConfig struct {
	SyncPlane syncplane.SyncPlaneSettings
	Telemetry telemetry.TelemetrySettings

	ProjectRepo projects.RedisProjectRepoConfig
	RuntimeRepo repo.RuntimeRepoSettings

	Storage struct {
		DataPath string
	}
}

func LoadConfig() (*ControllerConfig, error) {
	viper.SetConfigName("controller")
	viper.AddConfigPath("$PUPLOAD_CONFIG")
	viper.AddConfigPath("/etc/pupload")
	viper.AddConfigPath("$HOME/.config/pupload")
	viper.AddConfigPath("$HOME/.pupload")

	setDefaultConfig()

	var fileLookupError viper.ConfigFileNotFoundError
	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &fileLookupError) {
			return nil, err
		}
	}

	var cfg ControllerConfig
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil

}

func setDefaultConfig() {
	viper.SetDefault("SyncPlane.SelectedSyncPlane", "redis")
	viper.SetDefault("SyncPlane.Redis.Address", "localhost:6379")
	viper.SetDefault("SyncPlane.Redis.PoolSize", 10)
	viper.SetDefault("SyncPlane.Redis.MaxRetries", 3)
	viper.SetDefault("SyncPlane.ControllerStepInterval", "@every 1s")

	viper.SetDefault("ProjectRepo.Address", "localhost:6379")

	viper.SetDefault("RuntimeRepo.Type", "redis")
	viper.SetDefault("RuntimeRepo.Redis.Address", "localhost:6379")
}
