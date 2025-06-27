package hashcracktask

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
)

type repo struct {
	client     *mongo.Client
	wc         *writeconcern.WriteConcern
	rc         *readconcern.ReadConcern
	collection *mongo.Collection
	view       *mongo.Collection
	logger     zerolog.Logger
}

func NewRepo(logger zerolog.Logger, client *mongo.Client, cfg config.MongoDBConfig) repository.HashCrackTask {
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
	view := client.
		Database(cfg.DB).
		Collection(
			"hash_crack_tasks_with_subtasks",
			options.
				Collection().
				SetReadConcern(rc),
		)

	return &repo{
		client:     client,
		wc:         wc,
		rc:         rc,
		collection: collection,
		view:       view,
		logger: logger.With().
			Str("repo", "hash-crack-task").
			Str("type", "mongo").
			Logger(),
	}
}

func (r *repo) WithTransaction(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	r.logger.Debug().Msg("with transaction")

	txOpts := options.
		Transaction().
		SetReadConcern(r.rc).
		SetWriteConcern(r.wc)

	session, err := r.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	sessionFn := func(sessionContext mongo.SessionContext) (any, error) { //nolint:contextcheck
		return fn(sessionContext)
	}

	result, err := session.WithTransaction(ctx, sessionFn, txOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to execute with transaction: %w", err)
	}

	return result, nil
}

func (r *repo) GetAll(ctx context.Context, limit, offset int, withSubtasks bool) (
	[]*entity.HashCrackTaskWithSubtasks, error,
) {
	r.logger.Debug().
		Int("limit", limit).
		Int("offset", offset).
		Bool("with-subtasks", withSubtasks).
		Msg("get all")

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"createdAt": 1})

	return r.findAll(ctx, bson.M{}, withSubtasks, opts)
}

func (r *repo) CountAll(ctx context.Context) (int64, error) {
	r.logger.Debug().Msg("count all")

	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

func (r *repo) GetByHashAndMaxLength(
	ctx context.Context, hash string, maxLength int, withSubtasks bool,
) (*entity.HashCrackTaskWithSubtasks, error) {

	r.logger.Debug().
		Str("hash", hash).
		Int("max-length", maxLength).
		Bool("with-subtasks", withSubtasks).
		Msg("get by hash and max length")

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

	return r.findOne(ctx, filter, withSubtasks, options.FindOne())
}

func (r *repo) GetAllFinished(ctx context.Context, withSubtasks bool) ([]*entity.HashCrackTaskWithSubtasks, error) {
	r.logger.Debug().
		Bool("with-subtasks", withSubtasks).
		Msg("get all finished crack tasks")

	filter := bson.M{
		"$and": []bson.M{
			{"status": entity.HashCrackTaskStatusInProgress},
			{"finishedAt": bson.M{"$ne": nil}},
			{"finishedAt": bson.M{"$lt": time.Now()}},
		},
	}
	opts := options.Find().SetSort(bson.M{"createdAt": 1})

	return r.findAll(ctx, filter, withSubtasks, opts)
}

func (r *repo) GetAllExpired(
	ctx context.Context, maxAge time.Duration, withSubtasks bool,
) ([]*entity.HashCrackTaskWithSubtasks, error) {
	r.logger.Debug().
		Dur("max-age", maxAge).
		Bool("with-subtasks", withSubtasks).
		Msg("get all expired crack tasks")

	expirationTime := time.Now().Add(-maxAge)
	filter := bson.M{"createdAt": bson.M{"$lt": expirationTime}}

	return r.findAll(ctx, filter, withSubtasks, options.Find())
}

func (r *repo) Get(
	ctx context.Context, id primitive.ObjectID, withSubtasks bool,
) (*entity.HashCrackTaskWithSubtasks, error) {
	r.logger.Debug().
		Str("id", id.Hex()).
		Bool("with-subtasks", withSubtasks).
		Msg("get crack task")

	filter := bson.M{"_id": id}

	collection := r.collection
	if withSubtasks {
		collection = r.view
	}

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, repository.ErrCrackTaskNotFound
		}
		return nil, fmt.Errorf("failed to find one document: %w", result.Err())
	}

	var task entity.HashCrackTaskWithSubtasks
	if err := result.Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	return &task, nil
}

func (r *repo) Create(ctx context.Context, task *entity.HashCrackTask) error {
	r.logger.Debug().Str("id", task.ObjectID.Hex()).Msg("create task")

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

func (r *repo) DeleteAllByIDs(ctx context.Context, ids []primitive.ObjectID) error {
	r.logger.Debug().
		Int("count", len(ids)).
		Msg("delete crack tasks by ids")

	filter := bson.M{"_id": bson.M{"$in": ids}}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}

func (r *repo) findOne(
	ctx context.Context, filter interface{}, withSubtasks bool, opts ...*options.FindOneOptions,
) (*entity.HashCrackTaskWithSubtasks, error) {
	collection := r.collection
	if withSubtasks {
		collection = r.view
	}

	result := collection.FindOne(ctx, filter, opts...)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, repository.ErrCrackTaskNotFound
		}
		return nil, fmt.Errorf("failed to find one document: %w", result.Err())
	}

	var task entity.HashCrackTaskWithSubtasks
	if err := result.Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	return &task, nil
}

func (r *repo) findAll(
	ctx context.Context, filter interface{}, withSubtasks bool, opts ...*options.FindOptions,
) ([]*entity.HashCrackTaskWithSubtasks, error) {
	collection := r.collection
	if withSubtasks {
		collection = r.view
	}

	cursor, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}(cursor, ctx)

	var tasks []*entity.HashCrackTaskWithSubtasks
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return tasks, nil
}
