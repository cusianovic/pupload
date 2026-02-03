package config

import (
	"errors"

	"github.com/pupload/pupload/internal/resources"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
	"github.com/spf13/viper"
)

type WorkerConfig struct {
	Worker    WorkerSettings
	SyncPlane syncplane.SyncPlaneSettings
	Telemetry telemetry.TelemetrySettings
	Resources resources.ResourceSettings
	Runtime   RuntimeSettings

	Logging  LoggingSettings
	Security SecuritySettings
}

type WorkerSettings struct {
	ID string
}

type RuntimeSettings struct {
	ContainerEngine  string `json:"container_engine"`  // docker, podman, runsc, auto
	ContainerRuntime string `json:"container_runtime"` // runc, gvisor, auto

	EnableGPUSupport bool `json:"enable_gpu_support"`

	Gvisor GvisorSettings `json:"gvisor"`
}

type GvisorSettings struct {
	Platform string `json:"platform"` // systrap, kvm, ptrace
}

type LoggingSettings struct {
	LogLevel string `json:"log_level"` // debug, info, warn, error

}

type SecuritySettings struct {
	AllowedRegistries []string `json:"allowed_registries"` // all if empty
	ImageAllowlist    []string `json:"image_allowlist"`    //
	ImageBlocklist    []string `json:"image_blocklist"`
}

func LoadConfig() (*WorkerConfig, error) {
	viper.SetConfigName("worker")
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

	var cfg WorkerConfig
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

	viper.SetDefault("Resources.MaxCPU", "auto")
	viper.SetDefault("Resources.MaxMemory", "auto")
	viper.SetDefault("Resources.MaxStorage", "auto")

	viper.SetDefault("Runtime.ContainerEngine", "auto")
	viper.SetDefault("Runtime.ContainerRuntime", "auto")
	viper.SetDefault("Runtime.EnableGPUSupport", false)

	viper.SetDefault("Logging.LogLevel", "warn")

	viper.SetDefault("Security.AllowedRegistries", []string{
		"docker.io",
		"ghcr.io",
		"gcr.io",
		"public.ecr.aws",
	})

}
