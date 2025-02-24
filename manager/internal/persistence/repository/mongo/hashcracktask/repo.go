package hashcracktask

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"

	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
)

type repo struct {
	client     *mongo.Client
	collection *mongo.Collection
	logger     zerolog.Logger
}

func NewRepo(client *mongo.Client, cfg config.MongoDBConfig) repository.HashCrackTask {
	wc := &writeconcern.WriteConcern{
		W:       cfg.WriteConcern.W,
		Journal: cfg.WriteConcern.Journal,
	}
	rc := &readconcern.ReadConcern{
		Level: cfg.ReadConcern.Level,
	}
	collection := client.
		Database(cfg.DB).
		Collection(
			"hash_crack_tasks",
			options.
				Collection().
				SetReadConcern(rc).
				SetWriteConcern(wc),
		)

	return &repo{
		client:     client,
		collection: collection,
		logger: log.With().
			Str("repo", "hash-crack").
			Str("type", "mongo").
			Logger(),
	}
}

func (r *repo) GetAllByHashAndMaxLength(ctx context.Context, hash string, maxLength int) (
	[]*entity.HashCrackTask, error,
) {
	r.logger.Debug().
		Str("hash", hash).
		Int("max-length", maxLength).
		Msg("get all by hash and max length")

	filter := bson.M{
		"$and": []bson.M{
			{"hash": hash},
			{"maxLength": maxLength},
			{
				"$or": []bson.M{
					{"status": entity.HashCrackTaskStatusInProgress},
					{"status": entity.HashCrackTaskStatusReady},
				},
			},
		},
	}
	opts := options.Find().SetSort(bson.M{"createdAt": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}(cursor, ctx)

	var tasks []*entity.HashCrackTask
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return tasks, nil
}

func (r *repo) CountByStatus(ctx context.Context, status entity.HashCrackTaskStatus) (int, error) {
	r.logger.Debug().
		Str("status", status.String()).
		Msg("count crack tasks by status")

	filter := bson.M{"status": status}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return int(count), nil
}

func (r *repo) GetAllFinished(ctx context.Context) ([]*entity.HashCrackTask, error) {
	r.logger.Debug().Msg("get all finished crack tasks")

	filter := bson.M{
		"$and": []bson.M{
			{"status": entity.HashCrackTaskStatusInProgress},
			{"finishedAt": bson.M{"$ne": nil}},
			{"finishedAt": bson.M{"$lt": time.Now()}},
		},
	}
	opts := options.Find().SetSort(bson.M{"createdAt": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}(cursor, ctx)

	var tasks []*entity.HashCrackTask
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return tasks, nil
}

func (r *repo) Get(ctx context.Context, id bson.ObjectID) (*entity.HashCrackTask, error) {
	r.logger.Debug().Str("id", id.Hex()).Msg("get crack task")

	filter := bson.M{"_id": id}

	result := r.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, repository.ErrCrackTaskNotFound
		}
		return nil, fmt.Errorf("failed to find one document: %w", result.Err())
	}

	var task entity.HashCrackTask
	if err := result.Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	return &task, nil
}

func (r *repo) Create(ctx context.Context, task *entity.HashCrackTask) error {
	r.logger.Debug().Str("id", task.ObjectID.Hex()).Msg("insert crack task")

	_, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return repository.ErrCrackTaskExists
		}
		return fmt.Errorf("failed to insert one document: %w", err)
	}

	return nil
}

func (r *repo) Update(ctx context.Context, task *entity.HashCrackTask) error {
	r.logger.Debug().Str("id", task.ObjectID.Hex()).Msg("update crack task")

	filter := bson.M{"_id": task.ObjectID}
	update := bson.M{"$set": task}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update one document: %w", err)
	}

	if result.MatchedCount == 0 {
		return repository.ErrCrackTaskNotFound
	}

	return nil
}

func (r *repo) DeleteAllExpired(ctx context.Context, maxAge time.Duration) error {
	r.logger.Debug().
		Dur("max-age", maxAge).
		Msg("delete all expired crack tasks")

	expirationTime := time.Now().Add(-maxAge)
	filter := bson.M{"createdAt": bson.M{"$lt": expirationTime}}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}
