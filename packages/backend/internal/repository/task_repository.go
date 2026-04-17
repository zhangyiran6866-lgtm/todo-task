package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"todotask/backend/internal/model"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskRepository interface {
	InsertOne(ctx context.Context, task *model.Task) error
	FindMany(ctx context.Context, filter bson.D, limit int64) ([]*model.Task, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.Task, error)
	UpdateByID(ctx context.Context, id bson.ObjectID, update bson.M) error
	SoftDelete(ctx context.Context, id bson.ObjectID) error
}

type taskRepository struct {
	collection *mongo.Collection
}

func NewTaskRepository(db *mongo.Database) TaskRepository {
	return &taskRepository{
		collection: db.Collection("tasks"),
	}
}

func (r *taskRepository) InsertOne(ctx context.Context, task *model.Task) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	res, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		task.ID = oid
	}
	return nil
}

func (r *taskRepository) FindMany(ctx context.Context, filter bson.D, limit int64) ([]*model.Task, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	// 游标分页排序规范：先按 created_at DESC，再按 _id DESC 兜底
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []*model.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.Task, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	// 不查出已被软删除的数据
	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "is_deleted", Value: bson.M{"$ne": true}},
	}
	var task model.Task
	err := r.collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) UpdateByID(ctx context.Context, id bson.ObjectID, update bson.M) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "is_deleted", Value: bson.M{"$ne": true}},
	}

	res, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *taskRepository) SoftDelete(ctx context.Context, id bson.ObjectID) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "is_deleted", Value: bson.M{"$ne": true}},
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"deleted_at": time.Now(),
		},
	}
	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrTaskNotFound
	}
	return nil
}
