package service

import (
	"context"
	"todotask/backend/internal/model"
	"todotask/backend/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserService interface {
	GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}
