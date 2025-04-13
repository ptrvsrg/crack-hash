package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	URI      string
	Username string
	Password string
}

func NewClient(ctx context.Context, cfg Config) (*mongo.Client, error) {
	creds := options.Credential{
		Username: cfg.Username,
		Password: cfg.Password,
	}

	logOpts := options.
		Logger().
		SetSink(&logger{}).
		SetComponentLevel(options.LogComponentCommand, options.LogLevelDebug).
		SetComponentLevel(options.LogComponentConnection, options.LogLevelDebug)

	bsonOpts := &options.BSONOptions{
		UseJSONStructTags: true,
		NilSliceAsEmpty:   true,
	}

	opts := options.
		Client().
		ApplyURI(cfg.URI).
		SetAuth(creds).
		SetLoggerOptions(logOpts).
		SetBSONOptions(bsonOpts).
		SetCompressors([]string{"snappy", "zlib", "zstd"})

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := Ping(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

func Ping(ctx context.Context, client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return nil
}
