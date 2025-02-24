package helper

import "os"

func GetEnvOrDefault(key, defaultValue string) string {
	configPath, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return configPath
}
