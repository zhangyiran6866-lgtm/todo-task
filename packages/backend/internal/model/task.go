package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Task struct {
	ID          bson.ObjectID `bson:"_id,omitempty"  json:"id"`
	UserID      bson.ObjectID `bson:"user_id"        json:"user_id"`
	Title       string             `bson:"title"          json:"title"`
	Description string             `bson:"description"    json:"description"`
	Status      string             `bson:"status"         json:"status"`   // "todo"|"in_progress"|"done"
	Priority    string             `bson:"priority"       json:"priority"` // "low"|"important"|"urgent"|"critical"|"routine"
	DueAt       *time.Time         `bson:"due_at"         json:"due_at"`
	IsDeleted   bool               `bson:"is_deleted"     json:"-"`
	DeletedAt   *time.Time         `bson:"deleted_at"     json:"-"`
	CreatedAt   time.Time          `bson:"created_at"     json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"     json:"updated_at"`
}
