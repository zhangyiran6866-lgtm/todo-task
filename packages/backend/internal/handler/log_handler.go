package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "todotask/backend/internal/model"
	"todotask/backend/internal/repository"
	"todotask/backend/internal/service"
	applog "todotask/backend/pkg/logger"
	"todotask/backend/pkg/response"
)

type LogHandler struct {
	svc service.LogService
	log *zap.Logger
}

func NewLogHandler(svc service.LogService, log *zap.Logger) *LogHandler {
	return &LogHandler{svc: svc, log: log}
}

func (h *LogHandler) reqLogger(c *gin.Context, action string) *zap.Logger {
	return applog.WithContext(h.log, c.Request.Context()).With(
		zap.String(applog.FieldModule, "logs"),
		zap.String(applog.FieldAction, action),
	)
}

// ListLogs godoc
// @Summary 查询日志列表
// @Description Query structured logs with channel/level/module/keyword/time range filters
// @Tags Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channel query string false "Log channel: app|error|audit"
// @Param level query string false "Log level: debug|info|warn|error"
// @Param module query string false "Log module"
// @Param keyword query string false "Keyword full text search"
// @Param start_at query string false "Start time (RFC3339 or yyyy-mm-dd)"
// @Param end_at query string false "End time (RFC3339 or yyyy-mm-dd)"
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 20, max 100)"
// @Param limit query int false "Deprecated: alias of page_size"
// @Param cursor query string false "Deprecated: offset cursor"
// @Success 200 {object} response.Response{data=service.ListLogsResp} "Successfully retrieved logs"
// @Failure 400 {object} response.Response "Invalid query parameters"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /logs [get]
func (h *LogHandler) ListLogs(c *gin.Context) {
	reqLogger := h.reqLogger(c, "list_logs")

	var req service.ListLogsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		reqLogger.Warn("list logs bind failed", zap.Error(err))
		response.BadRequest(c, "查询参数不合法")
		return
	}

	respData, err := h.svc.ListLogs(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLogQuery) {
			response.BadRequest(c, "日志查询参数不合法")
			return
		}
		reqLogger.Error("list logs failed", zap.Error(err))
		response.InternalError(c, "查询日志失败")
		return
	}

	response.OK(c, respData)
}

// GetLog godoc
// @Summary 获取日志详情
// @Description Get full structured log content by id
// @Tags Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Log ID"
// @Param channel query string false "Log channel: app|error|audit"
// @Success 200 {object} response.Response{data=model.LogEntry} "Successfully retrieved log detail"
// @Failure 400 {object} response.Response "Invalid log id"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "Log not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /logs/{id} [get]
func (h *LogHandler) GetLog(c *gin.Context) {
	reqLogger := h.reqLogger(c, "get_log")

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.BadRequest(c, "日志 ID 不能为空")
		return
	}

	channel := c.Query("channel")
	entry, err := h.svc.GetLogByID(c.Request.Context(), id, channel)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLogQuery) {
			response.BadRequest(c, "日志参数不合法")
			return
		}
		if errors.Is(err, repository.ErrLogNotFound) {
			response.NotFound(c, "日志不存在")
			return
		}
		reqLogger.Error("get log failed", zap.Error(err))
		response.InternalError(c, "获取日志详情失败")
		return
	}

	response.OK(c, entry)
}
