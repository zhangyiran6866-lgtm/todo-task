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

func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	res, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailConflict) {
			response.Conflict(c, err.Error())
			return
		}
		h.log.Error("register failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.OK(c, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	res, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLogin) {
			response.Unauthorized(c, err.Error())
			return
		}
		h.log.Error("login failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.OK(c, res)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid refresh token")
		return
	}

	res, err := h.svc.Refresh(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.Unauthorized(c, err.Error())
			return
		}
		h.log.Error("refresh failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.OK(c, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid refresh token")
		return
	}

	if err := h.svc.Logout(c.Request.Context(), &req); err != nil {
		h.log.Error("logout failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.OK(c, nil)
}
