package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"todotask/backend/internal/model"
	"todotask/backend/internal/repository"
	"todotask/backend/pkg/hash"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrPasswordSame    = errors.New("new password must be different from old password")
	ErrEmptyProfile    = errors.New("at least one profile field is required")
	ErrInvalidProfile  = errors.New("invalid profile field")
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type UpdateProfileRequest struct {
	Nickname *string `json:"nickname,omitempty"`
	Language *string `json:"language,omitempty"`
	Theme    *string `json:"theme,omitempty"`
}

type UserService interface {
	GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	UpdateProfile(ctx context.Context, id bson.ObjectID, req *UpdateProfileRequest) (*model.User, error)
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

func (s *userService) UpdateProfile(ctx context.Context, id bson.ObjectID, req *UpdateProfileRequest) (*model.User, error) {
	update := bson.M{
		"updated_at": time.Now(),
	}

	if req.Nickname != nil {
		nickname := strings.TrimSpace(*req.Nickname)
		if nickname == "" || len([]rune(nickname)) > 30 {
			return nil, ErrInvalidProfile
		}
		update["nickname"] = nickname
	}
	if req.Language != nil {
		if !isValidLanguage(*req.Language) {
			return nil, ErrInvalidProfile
		}
		update["language"] = *req.Language
	}
	if req.Theme != nil {
		if !isValidTheme(*req.Theme) {
			return nil, ErrInvalidProfile
		}
		update["theme"] = *req.Theme
	}

	if len(update) == 1 {
		return nil, ErrEmptyProfile
	}

	if err := s.repo.UpdateByID(ctx, id, update); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func isValidLanguage(language string) bool {
	return language == "zh" || language == "en"
}

func isValidTheme(theme string) bool {
	switch theme {
	case "cyan", "purple", "green", "pink":
		return true
	default:
		return false
	}
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
