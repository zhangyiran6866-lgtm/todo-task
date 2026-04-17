package repository

import (
	"context"
	"time"
)

const dbOpTimeout = 5 * time.Second

func withDBTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, dbOpTimeout)
}
