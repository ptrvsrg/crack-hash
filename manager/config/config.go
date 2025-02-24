package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/ptrvsrg/crack-hash/manager/internal/helper"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/infrastructure/tasksplit/factory"
	"github.com/rs/zerolog/log"
)

type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Worker WorkerConfig `yaml:"worker"`
	Task   TaskConfig   `yaml:"task"`
}

type ServerConfig struct {
	Env  Env `yaml:"env" env:"SERVER_ENV" env-default:"dev" validate:"oneof=dev prod"`
	Port int `yaml:"port" env:"SERVER_PORT" env-default:"8080" validate:"required,min=-1,max=65535"`
}

type WorkerConfig struct {
	Address string `yaml:"address" env:"WORKER_ADDRESS" validate:"required,hostname_port"`
}

type TaskConfig struct {
	SplitStrategy factory.Strategy `yaml:"splitStrategy" env:"TASK_SPLIT_STRATEGY" env-default:"chunk-based" validate:"oneof=chunk-based"`
}

func LoadOrDie() Config {
	cfg, err := Load()
	if err != nil {
		log.Fatal().Err(err).Stack().Msgf("failed to load config: %v", err)
	}

	return cfg
}

func Load() (Config, error) {
	var cfg Config
	configName := getConfigName()

	if err := cleanenv.ReadConfig(configName, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to load config: %w", err)
	}

	if err := validator.New().Struct(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func getConfigName() string {
	return helper.GetEnvOrDefault("CONFIG_FILE", "config/config.yaml")
}
