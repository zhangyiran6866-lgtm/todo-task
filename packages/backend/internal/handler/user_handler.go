package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"todotask/backend/internal/middleware"
	_ "todotask/backend/internal/model"
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

func (h *UserHandler) getCurrentUserID(c *gin.Context) (bson.ObjectID, bool) {
	userIdV, exists := c.Get(middleware.CtxUserIDKey)
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return bson.NilObjectID, false
	}

	idStr, ok := userIdV.(string)
	if !ok {
		h.log.Error("invalid user_id type in context", zap.Any("type", userIdV))
		response.InternalError(c, "服务器内部错误")
		return bson.NilObjectID, false
	}

	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		h.log.Error("invalid user_id format in context", zap.String("id", idStr), zap.Error(err))
		response.InternalError(c, "服务器内部错误")
		return bson.NilObjectID, false
	}

	return id, true
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
	id, ok := h.getCurrentUserID(c)
	if !ok {
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		h.log.Error("get me failed", zap.Error(err))
		response.InternalError(c, "服务器内部错误")
		return
	}

	response.OK(c, user)
}

// UpdateMe godoc
// @Summary 更新当前用户信息
// @Description Update the authenticated user's profile fields
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.UpdateProfileRequest true "Profile update payload"
// @Success 200 {object} response.Response{data=model.User} "Successfully updated user profile"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/me [patch]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	id, ok := h.getCurrentUserID(c)
	if !ok {
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不合法")
		return
	}

	user, err := h.svc.UpdateProfile(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrEmptyProfile) {
			response.BadRequest(c, "至少需要更新一个字段")
			return
		}
		if errors.Is(err, service.ErrInvalidProfile) {
			response.BadRequest(c, "用户信息参数不合法")
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		h.log.Error("update me failed", zap.Error(err))
		response.InternalError(c, "服务器内部错误")
		return
	}

	response.OK(c, user)
}

// ChangePassword godoc
// @Summary 修改当前用户密码
// @Description Change the authenticated user's password after verifying old password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.ChangePasswordRequest true "Password change payload"
// @Success 200 {object} response.Response "Successfully changed password"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 401 {object} response.Response "Invalid old password"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	id, ok := h.getCurrentUserID(c)
	if !ok {
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不合法")
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), id, &req); err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			response.Unauthorized(c, "旧密码错误")
			return
		}
		if errors.Is(err, service.ErrPasswordSame) {
			response.BadRequest(c, "新密码不能与旧密码相同")
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		h.log.Error("change password failed", zap.Error(err))
		response.InternalError(c, "服务器内部错误")
		return
	}

	response.OK(c, nil)
}
