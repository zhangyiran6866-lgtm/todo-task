package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"todotask/backend/internal/service"
	applog "todotask/backend/pkg/logger"
	"todotask/backend/pkg/response"
)

type AuthHandler struct {
	svc service.AuthService
	log *zap.Logger
}

func NewAuthHandler(svc service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

func (h *AuthHandler) reqLogger(c *gin.Context, action string) *zap.Logger {
	return applog.WithContext(h.log, c.Request.Context()).With(
		zap.String(applog.FieldModule, "auth"),
		zap.String(applog.FieldAction, action),
	)
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
	reqLogger := h.reqLogger(c, "register")

	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Warn("register bind failed", zap.Error(err))
		response.BadRequest(c, "请求参数不合法")
		return
	}

	res, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailConflict) {
			applog.Audit(
				c.Request.Context(),
				"auth",
				"register_conflict",
				"user register conflict",
				zap.String("email", applog.MaskEmail(req.Email)),
			)
			response.Conflict(c, "邮箱已被注册")
			return
		}
		reqLogger.Error("register failed", zap.String("email", applog.MaskEmail(req.Email)), zap.Error(err))
		applog.Audit(
			c.Request.Context(),
			"auth",
			"register_failed",
			"user register failed",
			zap.String("email", applog.MaskEmail(req.Email)),
			zap.Error(err),
		)
		response.InternalError(c, "服务器内部错误")
		return
	}

	applog.Audit(
		c.Request.Context(),
		"auth",
		"register_success",
		"user registered",
		zap.String("email", applog.MaskEmail(req.Email)),
	)

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
	reqLogger := h.reqLogger(c, "login")

	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Warn("login bind failed", zap.Error(err))
		response.BadRequest(c, "请求参数不合法")
		return
	}

	res, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLogin) {
			applog.Audit(
				c.Request.Context(),
				"auth",
				"login_failed",
				"user login failed",
				zap.String("email", applog.MaskEmail(req.Email)),
			)
			response.Unauthorized(c, "邮箱或密码错误")
			return
		}
		reqLogger.Error("login failed", zap.String("email", applog.MaskEmail(req.Email)), zap.Error(err))
		applog.Audit(
			c.Request.Context(),
			"auth",
			"login_error",
			"user login error",
			zap.String("email", applog.MaskEmail(req.Email)),
			zap.Error(err),
		)
		response.InternalError(c, "服务器内部错误")
		return
	}

	applog.Audit(
		c.Request.Context(),
		"auth",
		"login_success",
		"user login success",
		zap.String("email", applog.MaskEmail(req.Email)),
	)

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
	reqLogger := h.reqLogger(c, "refresh")

	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Warn("refresh bind failed", zap.Error(err))
		response.BadRequest(c, "刷新令牌参数不合法")
		return
	}

	res, err := h.svc.Refresh(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.Unauthorized(c, "刷新令牌无效或已过期")
			return
		}
		reqLogger.Error("refresh failed", zap.String("refresh_token", applog.MaskToken(req.RefreshToken)), zap.Error(err))
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
	reqLogger := h.reqLogger(c, "logout")

	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Warn("logout bind failed", zap.Error(err))
		response.BadRequest(c, "刷新令牌参数不合法")
		return
	}

	if err := h.svc.Logout(c.Request.Context(), &req); err != nil {
		reqLogger.Error("logout failed", zap.String("refresh_token", applog.MaskToken(req.RefreshToken)), zap.Error(err))
		applog.Audit(
			c.Request.Context(),
			"auth",
			"logout_failed",
			"user logout failed",
			zap.Error(err),
		)
		response.InternalError(c, "服务器内部错误")
		return
	}

	applog.Audit(
		c.Request.Context(),
		"auth",
		"logout_success",
		"user logout success",
	)

	response.OK(c, nil)
}
