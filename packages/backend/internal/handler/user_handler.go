package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"todotask/backend/internal/middleware"
	"todotask/backend/internal/repository"
	"todotask/backend/internal/service"
	"todotask/backend/pkg/response"
)

type UserHandler struct {
	svc service.UserService
	log *zap.Logger
}

func NewUserHandler(svc service.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

// GetMe godoc
// @Summary 获取当前用户个人信息
// @Description Retrieve the currently authenticated user's profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=model.User} "Successfully retrieved user profile"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userIdV, exists := c.Get(middleware.CtxUserIDKey)
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	
	id, ok := userIdV.(bson.ObjectID)
	if !ok {
		response.InternalError(c, "internal server error")
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		h.log.Error("get me failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.OK(c, user)
}
