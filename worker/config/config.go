package config

import (
	_ "github.com/joho/godotenv/autoload"
)

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type (
	Env string

	Config struct {
		Server  ServerConfig  `yaml:"server"`
		Manager ManagerConfig `yaml:"manager"`
		Task    TaskConfig    `yaml:"task"`
	}

	ServerConfig struct {
		Env  Env `yaml:"env" env:"SERVER_ENV" env-default:"dev" validate:"oneof=dev prod"`
		Port int `yaml:"port" env:"SERVER_PORT" env-default:"8080" validate:"required,min=-1,max=65535"`
	}

	ManagerConfig struct {
		Address string `yaml:"address" env:"MANAGER_ADDRESS" validate:"required,http_url"`
	}

	TaskConfig struct {
		Split       TaskSplitConfig `yaml:"split"`
		Concurrency int             `yaml:"concurrency" env:"TASK_CONCURRENCY" env-default:"1000" validate:"min=1"`
	}

	TaskSplitConfig struct {
		Strategy  string `yaml:"strategy" env:"TASK_SPLIT_STRATEGY" env-default:"chunk-based" validate:"oneof=chunk-based"`
		ChunkSize int    `yaml:"chunkSize" env:"TASK_SPLIT_CHUNK_SIZE" env-default:"10000000" validate:"min=1"`
	}
)
