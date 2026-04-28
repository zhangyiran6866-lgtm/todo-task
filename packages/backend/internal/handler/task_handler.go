package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"todotask/backend/internal/middleware"
	_ "todotask/backend/internal/model"
	"todotask/backend/internal/repository"
	"todotask/backend/internal/service"
	applog "todotask/backend/pkg/logger"
	"todotask/backend/pkg/response"
)

type TaskHandler struct {
	svc    service.TaskService
	logger *zap.Logger
}

func NewTaskHandler(svc service.TaskService, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *TaskHandler) reqLogger(c *gin.Context, action string) *zap.Logger {
	return applog.WithContext(h.logger, c.Request.Context()).With(
		zap.String(applog.FieldModule, "task"),
		zap.String(applog.FieldAction, action),
	)
}

func (h *TaskHandler) getUID(c *gin.Context) (string, bool) {
	uid, exists := c.Get(middleware.CtxUserIDKey)
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return "", false
	}
	uidStr, ok := uid.(string)
	if !ok || uidStr == "" {
		h.reqLogger(c, "resolve_user").Error("invalid user id in context", zap.Any("user_id", uid))
		response.InternalError(c, "服务器内部错误")
		return "", false
	}
	return uidStr, true
}

// CreateTask godoc
// @Summary 创建任务
// @Description Create a new task for the authenticated user
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateTaskReq true "Task creation payload"
// @Success 200 {object} response.Response{data=model.Task} "Successfully created task"
// @Failure 400 {object} response.Response "Invalid task parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	reqLogger := h.reqLogger(c, "create_task")

	uid, ok := h.getUID(c)
	if !ok {
		return
	}

	var req service.CreateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Warn("create task bind error", zap.Error(err))
		response.BadRequest(c, "请求参数不合法")
		return
	}

	task, err := h.svc.CreateTask(c.Request.Context(), uid, &req)
	if err != nil {
		reqLogger.Error("create task failed", zap.Error(err))
		response.InternalError(c, "创建任务失败")
		return
	}

	response.OK(c, task)
}

// ListTasks godoc
// @Summary 获取任务列表
// @Description Get a paginated list of tasks for the authenticated user
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (todo, in_progress, done)"
// @Param priority query string false "Filter by priority (low, medium, high)"
// @Param limit query int false "Max number of items to return (default 20, max 50)"
// @Param cursor query string false "Cursor for pagination (ObjectId of the last item in previous page)"
// @Success 200 {object} response.Response{data=service.ListTasksResp} "Successfully retrieved tasks"
// @Failure 400 {object} response.Response "Invalid query parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	reqLogger := h.reqLogger(c, "list_tasks")

	uid, ok := h.getUID(c)
	if !ok {
		return
	}

	var req service.ListTasksReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "查询参数不合法")
		return
	}

	resp, err := h.svc.ListTasks(c.Request.Context(), uid, &req)
	if err != nil {
		reqLogger.Error("list tasks failed", zap.Error(err))
		response.InternalError(c, "获取任务列表失败")
		return
	}

	response.OK(c, resp)
}

// GetTask godoc
// @Summary 获取单个任务详情
// @Description Get details of a specific task by its ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} response.Response{data=model.Task} "Successfully retrieved task"
// @Failure 400 {object} response.Response "Invalid task ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden"
// @Failure 404 {object} response.Response "Task not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	reqLogger := h.reqLogger(c, "get_task")

	uid, ok := h.getUID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务 ID 不能为空")
		return
	}

	task, err := h.svc.GetTask(c.Request.Context(), uid, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			response.NotFound(c, "任务不存在")
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "无权限访问该任务")
			return
		}
		if err.Error() == "invalid task id" {
			response.BadRequest(c, "任务 ID 格式不正确")
			return
		}
		reqLogger.Error("get task failed", zap.Error(err))
		response.InternalError(c, "获取任务失败")
		return
	}

	response.OK(c, task)
}

// UpdateTask godoc
// @Summary 更新任务
// @Description Update fields of a specific task
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Param request body service.UpdateTaskReq true "Task update payload"
// @Success 200 {object} response.Response "Successfully updated task"
// @Failure 400 {object} response.Response "Invalid task ID or parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden"
// @Failure 404 {object} response.Response "Task not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /tasks/{id} [patch]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	reqLogger := h.reqLogger(c, "update_task")

	uid, ok := h.getUID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务 ID 不能为空")
		return
	}

	var req service.UpdateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Warn("update task bind error", zap.Error(err))
		response.BadRequest(c, "请求参数不合法")
		return
	}

	err := h.svc.UpdateTask(c.Request.Context(), uid, taskID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			response.NotFound(c, "任务不存在")
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "无权限更新该任务")
			return
		}
		if err.Error() == "invalid task id" {
			response.BadRequest(c, "任务 ID 格式不正确")
			return
		}
		reqLogger.Error("update task failed", zap.Error(err))
		response.InternalError(c, "更新任务失败")
		return
	}

	response.OK(c, nil)
}

// DeleteTask godoc
// @Summary 删除任务
// @Description Soft-delete a task by ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} response.Response "Successfully deleted task"
// @Failure 400 {object} response.Response "Invalid task ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden"
// @Failure 404 {object} response.Response "Task not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	reqLogger := h.reqLogger(c, "delete_task")

	uid, ok := h.getUID(c)
	if !ok {
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务 ID 不能为空")
		return
	}

	err := h.svc.DeleteTask(c.Request.Context(), uid, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			response.NotFound(c, "任务不存在")
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "无权限删除该任务")
			return
		}
		if err.Error() == "invalid task id" {
			response.BadRequest(c, "任务 ID 格式不正确")
			return
		}
		reqLogger.Error("delete task failed", zap.Error(err))
		applog.Audit(
			c.Request.Context(),
			"task",
			"delete_failed",
			"task delete failed",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		response.InternalError(c, "删除任务失败")
		return
	}

	applog.Audit(
		c.Request.Context(),
		"task",
		"delete_success",
		"task deleted",
		zap.String("task_id", taskID),
	)

	response.OK(c, nil)
}
