package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"todotask/backend/internal/model"
	"todotask/backend/internal/repository"
)

var (
	ErrForbidden            = errors.New("forbidden: you don't have permission to access this resource")
	ErrInvalidCursor        = errors.New("invalid cursor")
	ErrDueAtRequired        = errors.New("due_at is required")
	ErrInvalidTaskTimeRange = errors.New("start_at must be before or equal to due_at")
)

type CreateTaskReq struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	StartAt     *time.Time `json:"start_at"`
	DueAt       *time.Time `json:"due_at"`
}

type OptionalTime struct {
	Set   bool
	Value *time.Time
}

func (o *OptionalTime) UnmarshalJSON(data []byte) error {
	o.Set = true
	raw := strings.TrimSpace(string(data))
	if raw == "null" {
		o.Value = nil
		return nil
	}

	var parsed time.Time
	if err := parsed.UnmarshalJSON(data); err != nil {
		return err
	}
	o.Value = &parsed
	return nil
}

type UpdateTaskReq struct {
	Title       *string      `json:"title"`
	Status      *string      `json:"status"`
	Priority    *string      `json:"priority"`
	StartAt     OptionalTime `json:"start_at"`
	DueAt       OptionalTime `json:"due_at"`
	Description *string      `json:"description"`
}

type ListTasksReq struct {
	Status   string `form:"status"`
	Priority string `form:"priority"`
	Limit    int64  `form:"limit"`
	Cursor   string `form:"cursor"`
}

type ListTasksResp struct {
	Items      []*model.Task `json:"items"`
	NextCursor string        `json:"next_cursor"`
}

type TaskService interface {
	CreateTask(ctx context.Context, userID string, req *CreateTaskReq) (*model.Task, error)
	ListTasks(ctx context.Context, userID string, req *ListTasksReq) (*ListTasksResp, error)
	GetTask(ctx context.Context, userID string, taskID string) (*model.Task, error)
	UpdateTask(ctx context.Context, userID string, taskID string, req *UpdateTaskReq) error
	DeleteTask(ctx context.Context, userID string, taskID string) error
}

type taskService struct {
	repo repository.TaskRepository
}

type taskListCursorPayload struct {
	Version    int    `json:"v"`
	StatusRank int    `json:"status_rank"`
	StartAtNil int    `json:"start_at_nil"`
	StartAt    string `json:"start_at,omitempty"`
	DueAtNil   int    `json:"due_at_nil"`
	DueAt      string `json:"due_at,omitempty"`
	UpdatedAt  string `json:"updated_at"`
	ID         string `json:"id"`
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, userID string, req *CreateTaskReq) (*model.Task, error) {
	userOid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}
	if req.DueAt == nil {
		return nil, ErrDueAtRequired
	}
	if req.StartAt != nil && req.StartAt.After(*req.DueAt) {
		return nil, ErrInvalidTaskTimeRange
	}

	priority := req.Priority
	if priority == "" {
		priority = "low" // 默认为低优先级
	}

	now := time.Now()
	task := &model.Task{
		UserID:      userOid,
		Title:       req.Title,
		Description: req.Description,
		Status:      "todo", // 默认为 todo
		Priority:    priority,
		StartAt:     req.StartAt,
		DueAt:       req.DueAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = s.repo.InsertOne(ctx, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) ListTasks(ctx context.Context, userID string, req *ListTasksReq) (*ListTasksResp, error) {
	userOid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	filter := bson.D{
		{Key: "user_id", Value: userOid},
		{Key: "is_deleted", Value: bson.M{"$ne": true}},
	}

	if req.Status != "" {
		if req.Status == "expired" {
			filter = append(filter, bson.E{Key: "status", Value: bson.M{"$ne": "done"}})
			filter = append(filter, bson.E{Key: "due_at", Value: bson.M{"$lt": time.Now(), "$ne": nil}})
		} else {
			filter = append(filter, bson.E{Key: "status", Value: req.Status})
		}
	}
	if req.Priority != "" {
		filter = append(filter, bson.E{Key: "priority", Value: req.Priority})
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 50 {
		limit = 50
	}

	cursor, err := decodeTaskCursor(req.Cursor)
	if err != nil {
		return nil, err
	}

	entries, err := s.repo.FindManySorted(ctx, filter, limit+1, time.Now(), cursor)
	if err != nil {
		return nil, err
	}

	hasMore := int64(len(entries)) > limit
	if hasMore {
		entries = entries[:limit]
	}

	tasks := make([]*model.Task, 0, len(entries))
	for _, entry := range entries {
		tasks = append(tasks, entry.Task)
	}

	nextCursor := ""
	if hasMore && len(entries) > 0 {
		encoded, encodeErr := encodeTaskCursor(entries[len(entries)-1].Cursor)
		if encodeErr != nil {
			return nil, encodeErr
		}
		nextCursor = encoded
	}

	// Make sure Items is never nil pointer slice in json
	if tasks == nil {
		tasks = []*model.Task{}
	}

	return &ListTasksResp{
		Items:      tasks,
		NextCursor: nextCursor,
	}, nil
}

func (s *taskService) GetTask(ctx context.Context, userID string, taskID string) (*model.Task, error) {
	taskOid, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, errors.New("invalid task id")
	}

	task, err := s.repo.FindByID(ctx, taskOid)
	if err != nil {
		return nil, err
	}

	if task.UserID.Hex() != userID {
		return nil, ErrForbidden
	}

	return task, nil
}

func (s *taskService) UpdateTask(ctx context.Context, userID string, taskID string, req *UpdateTaskReq) error {
	taskOid, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return errors.New("invalid task id")
	}

	// 1. 先验证归属权
	task, err := s.repo.FindByID(ctx, taskOid)
	if err != nil {
		return err
	}
	if task.UserID.Hex() != userID {
		return ErrForbidden
	}
	if req.DueAt.Set && req.DueAt.Value == nil {
		return ErrDueAtRequired
	}

	resolvedDue := task.DueAt
	if req.DueAt.Set {
		resolvedDue = req.DueAt.Value
	}
	if resolvedDue == nil {
		return ErrDueAtRequired
	}

	resolvedStart := task.StartAt
	if req.StartAt.Set {
		resolvedStart = req.StartAt.Value
	}
	if resolvedStart != nil && resolvedStart.After(*resolvedDue) {
		return ErrInvalidTaskTimeRange
	}

	// 2. 组装部分更新字段
	update := bson.M{
		"updated_at": time.Now(),
	}
	if req.Title != nil {
		update["title"] = *req.Title
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Status != nil {
		update["status"] = *req.Status
	}
	if req.Priority != nil {
		update["priority"] = *req.Priority
	}
	if req.StartAt.Set {
		if req.StartAt.Value == nil {
			update["start_at"] = nil
		} else {
			update["start_at"] = *req.StartAt.Value
		}
	}
	if req.DueAt.Set && req.DueAt.Value != nil {
		update["due_at"] = *req.DueAt.Value
	}

	return s.repo.UpdateByID(ctx, taskOid, update)
}

func (s *taskService) DeleteTask(ctx context.Context, userID string, taskID string) error {
	taskOid, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return errors.New("invalid task id")
	}

	// 1. 验证归属权
	task, err := s.repo.FindByID(ctx, taskOid)
	if err != nil {
		return err
	}
	if task.UserID.Hex() != userID {
		return ErrForbidden
	}

	// 2. 执行软删除
	return s.repo.SoftDelete(ctx, taskOid)
}

func decodeTaskCursor(raw string) (*repository.TaskListCursor, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	var payload taskListCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, ErrInvalidCursor
	}
	if payload.Version != 2 {
		return nil, ErrInvalidCursor
	}

	oid, err := bson.ObjectIDFromHex(payload.ID)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, payload.UpdatedAt)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	var startAt *time.Time
	if payload.StartAt != "" {
		t, parseErr := time.Parse(time.RFC3339Nano, payload.StartAt)
		if parseErr != nil {
			return nil, ErrInvalidCursor
		}
		startAt = &t
	}

	var dueAt *time.Time
	if payload.DueAt != "" {
		t, parseErr := time.Parse(time.RFC3339Nano, payload.DueAt)
		if parseErr != nil {
			return nil, ErrInvalidCursor
		}
		dueAt = &t
	}

	return &repository.TaskListCursor{
		StatusRank: payload.StatusRank,
		StartAtNil: payload.StartAtNil,
		StartAt:    startAt,
		DueAtNil:   payload.DueAtNil,
		DueAt:      dueAt,
		UpdatedAt:  updatedAt,
		ID:         oid,
	}, nil
}

func encodeTaskCursor(cursor repository.TaskListCursor) (string, error) {
	payload := taskListCursorPayload{
		Version:    2,
		StatusRank: cursor.StatusRank,
		StartAtNil: cursor.StartAtNil,
		DueAtNil:   cursor.DueAtNil,
		UpdatedAt:  cursor.UpdatedAt.UTC().Format(time.RFC3339Nano),
		ID:         cursor.ID.Hex(),
	}
	if cursor.StartAt != nil {
		payload.StartAt = cursor.StartAt.UTC().Format(time.RFC3339Nano)
	}
	if cursor.DueAt != nil {
		payload.DueAt = cursor.DueAt.UTC().Format(time.RFC3339Nano)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
