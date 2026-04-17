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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	UpdateByID(ctx context.Context, id bson.ObjectID, update bson.M) error
	UpdatePassword(ctx context.Context, id bson.ObjectID, password string, updatedAt time.Time) error
}

type userRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	col := db.Collection("users")

	// Email 唯一索引
	idxCtx, idxCancel := withDBTimeout(context.Background())
	defer idxCancel()
	col.Indexes().CreateOne(idxCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// created_at 倒序索引
	idxCtx2, idxCancel2 := withDBTimeout(context.Background())
	defer idxCancel2()
	col.Indexes().CreateOne(idxCtx2, mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	})

	return &userRepository{col: col}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	res, err := r.col.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrUserAlreadyExists
		}
		return err
	}
	// MongoDB Driver v2 返回的 InsertedID 必须做类型断言
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		user.ID = id
	}
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var user model.User
	filter := bson.D{
		{Key: "email", Value: email},
		{Key: "is_deleted", Value: bson.D{{Key: "$ne", Value: true}}},
	}
	err := r.col.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var user model.User
	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "is_deleted", Value: bson.D{{Key: "$ne", Value: true}}},
	}
	err := r.col.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateByID(ctx context.Context, id bson.ObjectID, update bson.M) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "is_deleted", Value: bson.D{{Key: "$ne", Value: true}}},
	}

	res, err := r.col.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id bson.ObjectID, password string, updatedAt time.Time) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "is_deleted", Value: bson.D{{Key: "$ne", Value: true}}},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "password", Value: password},
			{Key: "updated_at", Value: updatedAt},
		}},
	}

	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}
