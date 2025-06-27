package config

import (
	"time"

	_ "github.com/joho/godotenv/autoload"
)

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type (
	Env string

	Config struct {
		Server ServerConfig
		AMQP   AMQPConfig
		Task   TaskConfig
	}

	ServerConfig struct {
		Env  Env `default:"dev" validate:"oneof=dev prod"`
		Port int `default:"8080" validate:"required,min=-1,max=65535"`
		Cors CorsConfig
	}

	CorsConfig struct {
		AllowedOrigins   []string      `default:"[\"*\"]"`
		AllowedMethods   []string      `default:"[\"GET\", \"POST\", \"PUT\", \"PATCH\", \"DELETE\", \"OPTIONS\"]"`
		AllowedHeaders   []string      `default:"[\"*\"]"`
		AllowCredentials bool          `default:"false"`
		MaxAge           time.Duration `default:"24h"`
	}

	AMQPConfig struct {
		URIs       []string `validate:"required,min=1,dive,required"`
		Username   string   `validate:"required"`
		Password   string   `validate:"required"`
		Prefetch   int      `default:"10" validate:"min=1"`
		Consumers  AMQPConsumersConfig
		Publishers AMQPPublishersConfig
	}

	AMQPConsumersConfig struct {
		TaskStarted AMQPConsumerConfig
	}

	AMQPConsumerConfig struct {
		Queue string `validate:"required"`
	}

	AMQPPublishersConfig struct {
		TaskResult AMQPPublisherConfig
	}

	AMQPPublisherConfig struct {
		Exchange   string `validate:"required"`
		RoutingKey string `validate:"required"`
	}

	TaskConfig struct {
		Split          TaskSplitConfig
		ProgressPeriod time.Duration `default:"5s" validate:"required"`
	}

	TaskSplitConfig struct {
		Strategy  string `default:"chunk-based" validate:"oneof=chunk-based"`
		ChunkSize int    `default:"10000000" validate:"min=1"`
	}
)
