package config

import (
	"pupload/internal/syncplane"
)

type WorkerConfig struct {
	Worker    WorkerSettings
	SyncPlane syncplane.SyncPlaneSettings
	Resources ResourceSettings
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

type ResourceSettings struct {
	MaxJobs int `json:"max_jobs"`

	MaxCPU     float32 // 0.5, 2, 1, etc.
	MaxMemory  string  // 1G, 512MB, etc.
	MaxStorage string  // 1G, 512MB, etc.
	MaxTimeout string  // 1H, 30s, 20m, etc.
}

type LoggingSettings struct {
	LogLevel string `json:"log_level"` // debug, info, warn, error

}

type SecuritySettings struct {
	AllowedRegistries []string `json:"allowed_registries"` // all if empty
	ImageAllowlist    []string `json:"image_allowlist"`    //
	ImageBlocklist    []string `json:"image_blocklist"`
}

func DefaultConfig() *WorkerConfig {
	return &WorkerConfig{
		Worker: WorkerSettings{
			ID: "0",
		},

		SyncPlane: syncplane.SyncPlaneSettings{
			SelectedSyncPlane: "redis",
			Redis: syncplane.RedisSettings{
				Address:  "localhost:6379",
				Password: "",
				DB:       0,

				PoolSize:   10,
				MaxRetries: 3,
			},
		},

		Resources: ResourceSettings{
			MaxJobs:    3,
			MaxCPU:     4.0,
			MaxMemory:  "8GB",
			MaxStorage: "50GB",
			MaxTimeout: "2h",
		},

		Runtime: RuntimeSettings{
			ContainerEngine:  "auto",
			ContainerRuntime: "auto",
			EnableGPUSupport: false,
		},

		Logging: LoggingSettings{
			LogLevel: "warn",
		},

		Security: SecuritySettings{
			AllowedRegistries: []string{
				"docker.io",
				"ghcr.io",
				"gcr.io",
				"public.ecr.aws",
			},

			ImageAllowlist: []string{},
			ImageBlocklist: []string{},
		},
	}
}
