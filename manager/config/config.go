package config

import (
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type (
	Config struct {
		Server ServerConfig `yaml:"server"`
		Worker WorkerConfig `yaml:"worker"`
		Task   TaskConfig   `yaml:"task"`
	}

	ServerConfig struct {
		Env  Env `yaml:"env" env:"SERVER_ENV" env-default:"dev" validate:"oneof=dev prod"`
		Port int `yaml:"port" env:"SERVER_PORT" env-default:"8080" validate:"required,min=-1,max=65535"`
	}

	WorkerConfig struct {
		Addresses []string           `yaml:"addresses" env:"WORKER_ADDRESSES" validate:"required,dive,http_url"`
		Health    WorkerHealthConfig `yaml:"health"`
	}

	WorkerHealthConfig struct {
		Path     string        `yaml:"path" env:"WORKER_HEALTH_PATH" validate:"required"`
		Interval time.Duration `yaml:"interval" env:"WORKER_HEALTH_INTERVAL" env-default:"1m"`
		Timeout  time.Duration `yaml:"timeout" env:"WORKER_HEALTH_TIMEOUT" env-default:"1m"`
		Retries  int           `yaml:"retries" env:"WORKER_HEALTH_RETRIES" env-default:"3"`
	}

	TaskConfig struct {
		Split       TaskSplitConfig `yaml:"split"`
		Timeout     time.Duration   `yaml:"timeout" env:"TASK_TIMEOUT" env-default:"1h"`
		Limit       int             `yaml:"limit" env:"TASK_LIMIT" env-default:"10" validate:"min=1"`
		MaxAge      time.Duration   `yaml:"maxAge" env:"TASK_MAX_AGE" env-default:"24h"`
		FinishDelay time.Duration   `yaml:"finishDelay" env:"TASK_FINISH_DELAY" env-default:"1m"`
	}

	TaskSplitConfig struct {
		Strategy  string `yaml:"strategy" env:"TASK_SPLIT_STRATEGY" env-default:"chunk-based" validate:"oneof=chunk-based"`
		ChunkSize int    `yaml:"chunkSize" env:"TASK_SPLIT_CHUNK_SIZE" env-default:"10000000" validate:"min=1"`
	}
)
