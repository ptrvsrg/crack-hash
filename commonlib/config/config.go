package config

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/num30/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func LoadOrDie[T any]() T {
	cfg, err := Load[T]()
	if err != nil {
		log.Fatal().Err(err).Stack().Msg("failed to load config")
	}

	return cfg
}

func Load[T any]() (T, error) {
	var cfg T
	configName := getConfigName()

	if err := config.NewConfReader(configName).Read(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

func getConfigName() string {
	configPath := getEnvOrDefault("CONFIG_FILE", "config/config.yaml")
	oldnew := make([]string, 2*len(viper.SupportedExts))
	for i, ext := range viper.SupportedExts {
		oldnew[2*i] = "." + ext
		oldnew[2*i+1] = ""
	}
	return strings.NewReplacer(oldnew...).Replace(configPath)
}

func getEnvOrDefault(key, defaultValue string) string {
	configPath, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return configPath
}
