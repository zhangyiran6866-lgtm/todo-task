package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"todotask/backend/internal/service"
	"todotask/backend/pkg/response"
)

type AuthHandler struct {
	svc service.AuthService
	log *zap.Logger
}

func NewAuthHandler(svc service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

// Register godoc
// @Summary 用户注册
// @Description Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Registration credentials"
// @Success 200 {object} response.Response{data=service.TokenResponse} "Successfully registered"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 409 {object} response.Response "Email already exists"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不合法")
		return
	}

	res, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailConflict) {
			response.Conflict(c, "邮箱已被注册")
			return
		}
		h.log.Error("register failed", zap.Error(err))
		response.InternalError(c, "服务器内部错误")
		return
	}

	response.OK(c, res)
}

// Login godoc
// @Summary 用户登录
// @Description Authenticate user and return access & refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=service.TokenResponse} "Successfully authenticated"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 401 {object} response.Response "Invalid login credentials"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不合法")
		return
	}

	res, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLogin) {
			response.Unauthorized(c, "邮箱或密码错误")
			return
		}
		h.log.Error("login failed", zap.Error(err))
		response.InternalError(c, "服务器内部错误")
		return
	}

	response.OK(c, res)
}

// Refresh godoc
// @Summary 刷新令牌
// @Description Obtain a new access token using a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body service.RefreshRequest true "Refresh token"
// @Success 200 {object} response.Response{data=service.TokenResponse} "Successfully refreshed token"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Invalid or expired token"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "刷新令牌参数不合法")
		return
	}

	res, err := h.svc.Refresh(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.Unauthorized(c, "刷新令牌无效或已过期")
			return
		}
		h.log.Error("refresh failed", zap.Error(err))
		response.InternalError(c, "服务器内部错误")
		return
	}

	response.OK(c, res)
}

// Logout godoc
// @Summary 用户登出
// @Description Invalidate the refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.RefreshRequest true "Refresh token to invalidate"
// @Success 200 {object} response.Response "Successfully logged out"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "刷新令牌参数不合法")
		return
	}

	if err := h.svc.Logout(c.Request.Context(), &req); err != nil {
		h.log.Error("logout failed", zap.Error(err))
		response.InternalError(c, "服务器内部错误")
		return
	}

	response.OK(c, nil)
}
