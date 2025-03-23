package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
)

func LoadOrDie[T any]() T {
	cfg, err := Load[T]()
	if err != nil {
		log.Fatal().Err(err).Stack().Msgf("failed to load config: %v", err)
	}

	return cfg
}

func Load[T any]() (T, error) {
	var cfg T
	configName := getConfigName()

	if err := cleanenv.ReadConfig(configName, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load config: %w", err)
	}

	if err := validator.New().Struct(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func getConfigName() string {
	return getEnvOrDefault("CONFIG_FILE", "config/config.yaml")
}

func getEnvOrDefault(key, defaultValue string) string {
	configPath, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return configPath
}
