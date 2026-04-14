package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"todotask/backend/internal/model"
	"todotask/backend/internal/repository"
)

var (
	ErrForbidden = errors.New("forbidden: you don't have permission to access this resource")
)

type CreateTaskReq struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	DueAt       *time.Time `json:"due_at"`
}

type UpdateTaskReq struct {
	Title       *string     `json:"title"`
	Status      *string     `json:"status"`
	Priority    *string     `json:"priority"`
	DueAt       *time.Time  `json:"due_at"`
	Description *string     `json:"description"`
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

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, userID string, req *CreateTaskReq) (*model.Task, error) {
	userOid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
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

	if req.Cursor != "" {
		cursorOid, err := bson.ObjectIDFromHex(req.Cursor)
		if err == nil {
			filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$lt": cursorOid}})
		}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 50 {
		limit = 50
	}

	tasks, err := s.repo.FindMany(ctx, filter, limit)
	if err != nil {
		return nil, err
	}

	nextCursor := ""
	if len(tasks) > 0 && int64(len(tasks)) == limit {
		nextCursor = tasks[len(tasks)-1].ID.Hex()
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
	if req.DueAt != nil {
		update["due_at"] = *req.DueAt
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
