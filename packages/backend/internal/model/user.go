package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User 表示用户信息
type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string        `bson:"email"         json:"email"`
	Password  string        `bson:"password"      json:"-"`
	Nickname  string        `bson:"nickname"      json:"nickname"`
	Language  string        `bson:"language"      json:"language"`
	Theme     string        `bson:"theme"         json:"theme"`
	IsDeleted bool          `bson:"is_deleted"    json:"-"`
	CreatedAt time.Time     `bson:"created_at"    json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"    json:"updated_at"`
}

// TokenBlacklist 表示 JWT Refresh Token 黑名单
type TokenBlacklist struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Token     string        `bson:"token"         json:"-"`
	ExpiredAt time.Time     `bson:"expired_at"    json:"-"`
	CreatedAt time.Time     `bson:"created_at"    json:"-"`
}
