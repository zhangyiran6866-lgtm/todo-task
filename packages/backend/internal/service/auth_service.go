package service

import (
	"context"
	"errors"
	"time"

	"todotask/backend/internal/model"
	"todotask/backend/internal/repository"
	"todotask/backend/pkg/config"
	"todotask/backend/pkg/hash"
	"todotask/backend/pkg/jwt"
)

var (
	ErrEmailConflict = errors.New("email already exists")
	ErrInvalidLogin  = errors.New("invalid email or password")
	ErrInvalidToken  = errors.New("invalid or expired token")
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error)
	Refresh(ctx context.Context, req *RefreshRequest) (*TokenResponse, error)
	Logout(ctx context.Context, req *RefreshRequest) error
}

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	jwtCfg    *config.JWTConfig
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, jwtCfg *config.JWTConfig) AuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtCfg:    jwtCfg,
	}
}

func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	hashedPassword, err := hash.MakePassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:     req.Email,
		Password:  hashedPassword,
		Nickname:  req.Nickname,
		Language:  "zh",   // Default
		Theme:     "cyan", // Default
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrEmailConflict
		}
		return nil, err
	}

	access, refresh, err := jwt.GenerateTokens(user.ID, s.jwtCfg)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidLogin
		}
		return nil, err
	}

	if !hash.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidLogin
	}

	access, refresh, err := jwt.GenerateTokens(user.ID, s.jwtCfg)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, req *RefreshRequest) (*TokenResponse, error) {
	isBlacklisted, err := s.tokenRepo.IsBlacklisted(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, ErrInvalidToken
	}

	claims, err := jwt.ParseToken(req.RefreshToken, s.jwtCfg.RefreshSecret)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Add old refresh token to blacklist
	_ = s.tokenRepo.AddToBlacklist(ctx, req.RefreshToken, claims.ExpiresAt.Time)

	// Issue new tokens
	access, refresh, err := jwt.GenerateTokens(claims.UserID, s.jwtCfg)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *authService) Logout(ctx context.Context, req *RefreshRequest) error {
	claims, err := jwt.ParseToken(req.RefreshToken, s.jwtCfg.RefreshSecret)
	if err != nil {
		return nil // Ignore expired/invalid token on logout
	}
	return s.tokenRepo.AddToBlacklist(ctx, req.RefreshToken, claims.ExpiresAt.Time)
}
