package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"todotask/backend/internal/model"
	"todotask/backend/internal/repository"
)

var (
	ErrInvalidLogQuery = errors.New("invalid log query")
)

type ListLogsReq struct {
	Channel  string `form:"channel"`
	Level    string `form:"level"`
	Module   string `form:"module"`
	Keyword  string `form:"keyword"`
	StartAt  string `form:"start_at"`
	EndAt    string `form:"end_at"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Limit    int    `form:"limit"`  // backward-compatible alias of page_size
	Cursor   string `form:"cursor"` // backward-compatible alias of offset pagination
}

type LogPagination struct {
	Total      int  `json:"total"`
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

type ListLogsResp struct {
	Items      []*model.LogEntry `json:"items"`
	Pagination LogPagination     `json:"pagination"`
	NextCursor string            `json:"next_cursor,omitempty"` // backward-compatible field
}

type LogService interface {
	ListLogs(ctx context.Context, req *ListLogsReq) (*ListLogsResp, error)
	GetLogByID(ctx context.Context, id string, channel string) (*model.LogEntry, error)
}

type logService struct {
	repo repository.LogRepository
}

func NewLogService(repo repository.LogRepository) LogService {
	return &logService{repo: repo}
}

func (s *logService) ListLogs(ctx context.Context, req *ListLogsReq) (*ListLogsResp, error) {
	startAt, endAt, err := parseLogTimeRange(req.StartAt, req.EndAt)
	if err != nil {
		return nil, err
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = req.Limit
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}

	page := req.Page
	offset := 0
	if page > 0 {
		offset = (page - 1) * pageSize
	} else {
		offset, err = parseCursor(req.Cursor)
		if err != nil {
			return nil, err
		}
		page = offset/pageSize + 1
	}

	if page <= 0 {
		page = 1
	}

	channel := strings.ToLower(strings.TrimSpace(req.Channel))
	if channel != "" && !isAllowedChannel(channel) {
		return nil, ErrInvalidLogQuery
	}

	level := strings.ToLower(strings.TrimSpace(req.Level))
	if level != "" && !isAllowedLevel(level) {
		return nil, ErrInvalidLogQuery
	}

	filter := repository.LogFilter{
		Channel: channel,
		Level:   level,
		Module:  strings.ToLower(strings.TrimSpace(req.Module)),
		Keyword: strings.TrimSpace(req.Keyword),
		StartAt: startAt,
		EndAt:   endAt,
	}

	items, total, err := s.repo.List(ctx, filter, offset, pageSize)
	if err != nil {
		return nil, err
	}

	nextCursor := ""
	nextOffset := offset + len(items)
	if nextOffset < total {
		nextCursor = strconv.Itoa(nextOffset)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &ListLogsResp{
		Items: items,
		Pagination: LogPagination{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		NextCursor: nextCursor,
	}, nil
}

func (s *logService) GetLogByID(ctx context.Context, id string, channel string) (*model.LogEntry, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidLogQuery
	}
	return s.repo.FindByID(ctx, id, strings.TrimSpace(strings.ToLower(channel)))
}

func parseCursor(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}

	offset, err := strconv.Atoi(raw)
	if err != nil || offset < 0 {
		return 0, ErrInvalidLogQuery
	}
	return offset, nil
}

func parseLogTimeRange(startRaw, endRaw string) (*time.Time, *time.Time, error) {
	start, err := parseLogTime(startRaw)
	if err != nil {
		return nil, nil, err
	}
	end, err := parseLogTime(endRaw)
	if err != nil {
		return nil, nil, err
	}

	if start != nil && end != nil && start.After(*end) {
		return nil, nil, ErrInvalidLogQuery
	}

	return start, end, nil
}

func parseLogTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return &t, nil
		}
	}
	return nil, ErrInvalidLogQuery
}

func isAllowedChannel(channel string) bool {
	switch channel {
	case "app", "error", "audit":
		return true
	default:
		return false
	}
}

func isAllowedLevel(level string) bool {
	switch level {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
