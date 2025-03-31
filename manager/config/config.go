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
		Server  ServerConfig
		MongoDB MongoDBConfig
		AMQP    AMQPConfig
		Task    TaskConfig
	}

	ServerConfig struct {
		Env  Env `default:"dev" validate:"required,oneof=dev prod"`
		Port int `default:"8080" validate:"required,min=-1,max=65535"`
	}

	MongoDBConfig struct {
		URI          string `validate:"required"`
		Username     string `validate:"required"`
		Password     string `validate:"required"`
		DB           string `validate:"required"`
		WriteConcern MongoDBWriteConcernConfig
		ReadConcern  MongoDBReadConcernConfig
	}

	MongoDBWriteConcernConfig struct {
		W       interface{} `default:"majority" validate:"required"`
		Journal *bool
	}

	MongoDBReadConcernConfig struct {
		Level string `default:"majority" validate:"required,oneof=local majority available linearizable snapshot"`
	}

	AMQPConfig struct {
		URIs       []string `validate:"required,min=1,dive,required"`
		Username   string   `validate:"required"`
		Password   string   `validate:"required"`
		Prefetch   int      `default:"20" validate:"required,min=1"`
		Consumers  AMQPConsumersConfig
		Publishers AMQPPublishersConfig
	}

	AMQPConsumersConfig struct {
		TaskResult AMQPConsumerConfig
	}

	AMQPConsumerConfig struct {
		Queue string `validate:"required"`
	}

	AMQPPublishersConfig struct {
		TaskStarted AMQPPublisherConfig
	}

	AMQPPublisherConfig struct {
		Exchange   string `validate:"required"`
		RoutingKey string `validate:"required"`
	}

	TaskConfig struct {
		Split       TaskSplitConfig
		Alphabet    string        `default:"abcdefghijklmnopqrstuvwxyz0123456789" validate:"required"`
		Timeout     time.Duration `default:"1h" validate:"required"`
		Limit       int           `default:"10" validate:"required,min=1"`
		MaxAge      time.Duration `default:"24h" validate:"required"`
		FinishDelay time.Duration `default:"1m" validate:"required"`
	}

	TaskSplitConfig struct {
		Strategy  string `default:"chunk-based" validate:"required,oneof=chunk-based"`
		ChunkSize int    `default:"10000000" validate:"required,min=1"`
	}
)
