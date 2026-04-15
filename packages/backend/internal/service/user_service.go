package service

import (
	"context"
	"errors"
	"time"

	"todotask/backend/internal/model"
	"todotask/backend/internal/repository"
	"todotask/backend/pkg/hash"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrPasswordSame    = errors.New("new password must be different from old password")
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type UserService interface {
	GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	ChangePassword(ctx context.Context, id bson.ObjectID, req *ChangePasswordRequest) error
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

func (s *userService) ChangePassword(ctx context.Context, id bson.ObjectID, req *ChangePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !hash.CheckPassword(req.OldPassword, user.Password) {
		return ErrInvalidPassword
	}
	if hash.CheckPassword(req.NewPassword, user.Password) {
		return ErrPasswordSame
	}

	hashedPassword, err := hash.MakePassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, id, hashedPassword, time.Now())
}
