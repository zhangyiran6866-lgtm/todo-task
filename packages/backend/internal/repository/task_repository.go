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
	FindManySorted(ctx context.Context, filter bson.D, limit int64, now time.Time, cursor *TaskListCursor) ([]TaskListItem, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.Task, error)
	UpdateByID(ctx context.Context, id bson.ObjectID, update bson.M) error
	SoftDelete(ctx context.Context, id bson.ObjectID) error
}

type TaskListCursor struct {
	StatusRank int
	StartAtNil int
	StartAt    *time.Time
	DueAtNil   int
	DueAt      *time.Time
	UpdatedAt  time.Time
	ID         bson.ObjectID
}

type TaskListItem struct {
	Task   *model.Task
	Cursor TaskListCursor
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

func (r *taskRepository) FindManySorted(ctx context.Context, filter bson.D, limit int64, now time.Time, cursor *TaskListCursor) ([]TaskListItem, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	addFields := bson.D{
		{
			Key: "status_rank",
			Value: bson.D{
				{
					Key: "$switch",
					Value: bson.D{
						{
							Key: "branches",
							Value: bson.A{
								bson.D{
									{
										Key: "case",
										Value: bson.D{
											{Key: "$eq", Value: bson.A{"$status", "done"}},
										},
									},
									{Key: "then", Value: 2},
								},
								bson.D{
									{
										Key: "case",
										Value: bson.D{
											{
												Key: "$and",
												Value: bson.A{
													bson.D{{Key: "$ne", Value: bson.A{"$status", "done"}}},
													bson.D{{Key: "$ne", Value: bson.A{"$due_at", nil}}},
													bson.D{{Key: "$lt", Value: bson.A{"$due_at", now}}},
												},
											},
										},
									},
									{Key: "then", Value: 1},
								},
							},
						},
						{Key: "default", Value: 0},
					},
				},
			},
		},
		{
			Key: "start_at_nil",
			Value: bson.D{
				{Key: "$cond", Value: bson.A{bson.D{{Key: "$eq", Value: bson.A{"$start_at", nil}}}, 1, 0}},
			},
		},
		{
			Key: "due_at_nil",
			Value: bson.D{
				{Key: "$cond", Value: bson.A{bson.D{{Key: "$eq", Value: bson.A{"$due_at", nil}}}, 1, 0}},
			},
		},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$addFields", Value: addFields}},
	}

	if cursor != nil {
		startAt := any(nil)
		if cursor.StartAt != nil {
			startAt = *cursor.StartAt
		}

		dueAt := any(nil)
		if cursor.DueAt != nil {
			dueAt = *cursor.DueAt
		}

		cursorMatch := bson.D{
			{
				Key: "$expr",
				Value: bson.D{
					{
						Key: "$or",
						Value: bson.A{
							bson.D{{Key: "$gt", Value: bson.A{"$status_rank", cursor.StatusRank}}},
							bson.D{{Key: "$and", Value: bson.A{
								bson.D{{Key: "$eq", Value: bson.A{"$status_rank", cursor.StatusRank}}},
								bson.D{{Key: "$gt", Value: bson.A{"$start_at_nil", cursor.StartAtNil}}},
							}}},
							bson.D{{Key: "$and", Value: bson.A{
								bson.D{{Key: "$eq", Value: bson.A{"$status_rank", cursor.StatusRank}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at_nil", cursor.StartAtNil}}},
								bson.D{{Key: "$gt", Value: bson.A{"$start_at", startAt}}},
							}}},
							bson.D{{Key: "$and", Value: bson.A{
								bson.D{{Key: "$eq", Value: bson.A{"$status_rank", cursor.StatusRank}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at_nil", cursor.StartAtNil}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at", startAt}}},
								bson.D{{Key: "$gt", Value: bson.A{"$due_at_nil", cursor.DueAtNil}}},
							}}},
							bson.D{{Key: "$and", Value: bson.A{
								bson.D{{Key: "$eq", Value: bson.A{"$status_rank", cursor.StatusRank}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at_nil", cursor.StartAtNil}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at", startAt}}},
								bson.D{{Key: "$eq", Value: bson.A{"$due_at_nil", cursor.DueAtNil}}},
								bson.D{{Key: "$gt", Value: bson.A{"$due_at", dueAt}}},
							}}},
							bson.D{{Key: "$and", Value: bson.A{
								bson.D{{Key: "$eq", Value: bson.A{"$status_rank", cursor.StatusRank}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at_nil", cursor.StartAtNil}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at", startAt}}},
								bson.D{{Key: "$eq", Value: bson.A{"$due_at_nil", cursor.DueAtNil}}},
								bson.D{{Key: "$eq", Value: bson.A{"$due_at", dueAt}}},
								bson.D{{Key: "$lt", Value: bson.A{"$updated_at", cursor.UpdatedAt}}},
							}}},
							bson.D{{Key: "$and", Value: bson.A{
								bson.D{{Key: "$eq", Value: bson.A{"$status_rank", cursor.StatusRank}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at_nil", cursor.StartAtNil}}},
								bson.D{{Key: "$eq", Value: bson.A{"$start_at", startAt}}},
								bson.D{{Key: "$eq", Value: bson.A{"$due_at_nil", cursor.DueAtNil}}},
								bson.D{{Key: "$eq", Value: bson.A{"$due_at", dueAt}}},
								bson.D{{Key: "$eq", Value: bson.A{"$updated_at", cursor.UpdatedAt}}},
								bson.D{{Key: "$lt", Value: bson.A{"$_id", cursor.ID}}},
							}}},
						},
					},
				},
			},
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: cursorMatch}})
	}

	pipeline = append(
		pipeline,
		bson.D{
			{
				Key: "$sort",
				Value: bson.D{
					{Key: "status_rank", Value: 1},
					{Key: "start_at_nil", Value: 1},
					{Key: "start_at", Value: 1},
					{Key: "due_at_nil", Value: 1},
					{Key: "due_at", Value: 1},
					{Key: "updated_at", Value: -1},
					{Key: "_id", Value: -1},
				},
			},
		},
		bson.D{{Key: "$limit", Value: limit}},
	)

	type taskListDoc struct {
		model.Task `bson:",inline"`
		StatusRank int `bson:"status_rank"`
		StartAtNil int `bson:"start_at_nil"`
		DueAtNil   int `bson:"due_at_nil"`
	}

	aggCursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer aggCursor.Close(ctx)

	var docs []taskListDoc
	if err := aggCursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	items := make([]TaskListItem, 0, len(docs))
	for _, doc := range docs {
		task := doc.Task
		items = append(items, TaskListItem{
			Task: &task,
			Cursor: TaskListCursor{
				StatusRank: doc.StatusRank,
				StartAtNil: doc.StartAtNil,
				StartAt:    task.StartAt,
				DueAtNil:   doc.DueAtNil,
				DueAt:      task.DueAt,
				UpdatedAt:  task.UpdatedAt,
				ID:         task.ID,
			},
		})
	}

	return items, nil
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
