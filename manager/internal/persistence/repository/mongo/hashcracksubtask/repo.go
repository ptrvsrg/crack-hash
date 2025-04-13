package hashcracksubtask

import (
	"context"
	"errors"
	"fmt"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/entity"
	"github.com/ptrvsrg/crack-hash/manager/internal/persistence/repository"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/multierr"
)

type repo struct {
	client     *mongo.Client
	rc         *readconcern.ReadConcern
	wc         *writeconcern.WriteConcern
	collection *mongo.Collection
	logger     zerolog.Logger
}

func NewRepo(logger zerolog.Logger, client *mongo.Client, cfg config.MongoDBConfig) repository.HashCrackSubtask {
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
			"hash_crack_subtasks",
			options.
				Collection().
				SetReadConcern(rc).
				SetWriteConcern(wc),
		)

	return &repo{
		client:     client,
		wc:         wc,
		rc:         rc,
		collection: collection,
		logger: logger.With().
			Str("repo", "hash-crack-subtask").
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

	sessionFn := func(sessionContext mongo.SessionContext) (any, error) {
		return fn(sessionContext)
	}

	result, err := session.WithTransaction(ctx, sessionFn, txOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to execute with transaction: %w", err)
	}

	return result, nil
}

func (r *repo) GetByTaskIDAndPartNumber(ctx context.Context, taskID primitive.ObjectID, partNumber int) (*entity.HashCrackSubtask, error) {
	r.logger.Debug().
		Str("task-id", taskID.Hex()).
		Int("part-number", partNumber).
		Msg("get subtask by task id and part number")

	filter := bson.M{"taskId": taskID, "partNumber": partNumber}

	return r.findOne(ctx, filter, options.FindOne())
}

func (r *repo) GetAllByTaskID(ctx context.Context, taskID primitive.ObjectID) ([]*entity.HashCrackSubtask, error) {
	r.logger.Debug().
		Str("task-id", taskID.Hex()).
		Msg("get subtasks by task id")

	filter := bson.M{"taskId": taskID}
	opts := options.Find().SetSort(bson.M{"createdAt": 1})

	return r.findAll(ctx, filter, opts)
}

func (r *repo) GetAllByTaskIDs(ctx context.Context, taskIDs []primitive.ObjectID) ([]*entity.HashCrackSubtask, error) {
	r.logger.Debug().
		Int("count", len(taskIDs)).
		Msg("get subtasks by task ids")

	filter := bson.M{"taskId": bson.M{"$in": taskIDs}}
	opts := options.Find().SetSort(bson.M{"createdAt": 1})

	return r.findAll(ctx, filter, opts)
}

func (r *repo) GetAllByStatus(ctx context.Context, status entity.HashCrackSubtaskStatus) ([]*entity.HashCrackSubtask, error) {
	r.logger.Debug().
		Str("status", status.String()).
		Msg("get subtasks by status")

	filter := bson.M{"status": status}
	opts := options.Find().SetSort(bson.M{"createdAt": 1})

	return r.findAll(ctx, filter, opts)
}

func (r *repo) Create(ctx context.Context, task *entity.HashCrackSubtask) error {
	r.logger.Debug().Msg("create subtask")

	_, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return repository.ErrCrackSubtaskExists
		}
		return fmt.Errorf("failed to insert one document: %w", err)
	}

	return nil
}

func (r *repo) CreateAll(ctx context.Context, tasks []*entity.HashCrackSubtask) error {
	ids := lo.Map(tasks, func(task *entity.HashCrackSubtask, _ int) string {
		return task.ObjectID.Hex()
	})

	r.logger.Debug().
		Int("count", len(tasks)).
		Strs("ids", ids).
		Msg("create subtasks")

	documents := lo.Map(tasks, func(task *entity.HashCrackSubtask, _ int) interface{} {
		return task
	})

	_, err := r.collection.InsertMany(ctx, documents)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return repository.ErrCrackSubtaskExists
		}
		return fmt.Errorf("failed to insert many documents: %w", err)
	}

	return nil
}

func (r *repo) Update(ctx context.Context, task *entity.HashCrackSubtask) error {
	r.logger.Debug().
		Str("id", task.ObjectID.Hex()).
		Msg("update subtask")

	filter := bson.M{"_id": task.ObjectID}
	update := bson.M{"$set": task}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update one document: %w", err)
	}

	return nil
}

func (r *repo) UpdateAll(ctx context.Context, tasks []*entity.HashCrackSubtask) error {
	ids := lo.Map(tasks, func(task *entity.HashCrackSubtask, _ int) string {
		return task.ObjectID.Hex()
	})

	r.logger.Debug().
		Int("count", len(tasks)).
		Strs("ids", ids).
		Msg("update subtasks")

	errs := make([]error, 0)
	for _, task := range tasks {
		filter := bson.M{"_id": task.ObjectID}
		update := bson.M{"$set": task}

		_, err := r.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return multierr.Combine(errs...)
}

func (r *repo) DeleteAllByIDs(ctx context.Context, ids []primitive.ObjectID) error {
	idRaws := lo.Map(ids, func(id primitive.ObjectID, _ int) string {
		return id.Hex()
	})

	r.logger.Debug().
		Int("count", len(ids)).
		Strs("ids", idRaws).
		Msg("delete subtasks by ids")

	filter := bson.M{"_id": bson.M{"$in": ids}}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}

func (r *repo) findOne(
	ctx context.Context, filter interface{}, opts ...*options.FindOneOptions,
) (*entity.HashCrackSubtask, error) {
	result := r.collection.FindOne(ctx, filter, opts...)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, repository.ErrCrackSubtaskNotFound
		}
		return nil, fmt.Errorf("failed to find one document: %w", result.Err())
	}

	var subtask entity.HashCrackSubtask
	if err := result.Decode(&subtask); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	return &subtask, nil
}

func (r *repo) findAll(
	ctx context.Context, filter interface{}, opts ...*options.FindOptions,
) ([]*entity.HashCrackSubtask, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}(cursor, ctx)

	var subtasks []*entity.HashCrackSubtask
	if err := cursor.All(ctx, &subtasks); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return subtasks, nil
}
