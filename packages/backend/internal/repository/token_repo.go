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

type TokenRepository interface {
	AddToBlacklist(ctx context.Context, token string, expiredAt time.Time) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

type tokenRepository struct {
	col *mongo.Collection
}

func NewTokenRepository(db *mongo.Database) TokenRepository {
	col := db.Collection("token_blacklist")

	// Token 唯一索引
	idxCtx, idxCancel := withDBTimeout(context.Background())
	defer idxCancel()
	col.Indexes().CreateOne(idxCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "token", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// TTL 索引，自动清理过期 Token (ExpireAfterSeconds=0 配合 Date 类型)
	idxCtx2, idxCancel2 := withDBTimeout(context.Background())
	defer idxCancel2()
	col.Indexes().CreateOne(idxCtx2, mongo.IndexModel{
		Keys:    bson.D{{Key: "expired_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})

	return &tokenRepository{col: col}
}

func (r *tokenRepository) AddToBlacklist(ctx context.Context, token string, expiredAt time.Time) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	doc := model.TokenBlacklist{
		ID:        bson.NewObjectID(),
		Token:     token,
		ExpiredAt: expiredAt,
		CreatedAt: time.Now(),
	}
	_, err := r.col.InsertOne(ctx, doc)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return err
	}
	return nil
}

func (r *tokenRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	filter := bson.D{{Key: "token", Value: token}}
	err := r.col.FindOne(ctx, filter).Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
