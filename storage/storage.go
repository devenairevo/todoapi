package storage

import (
	"context"
	"errors"
	"tasks/models"
)

type TaskStorage interface {
	CreateTask(ctx context.Context, task models.Task) (models.Task, error)
	GetTaskByID(ctx context.Context, id string) (models.Task, error)
	GetAllTasks(ctx context.Context) ([]models.Task, error)
	UpdateTask(ctx context.Context, task models.Task) (models.Task, error)
	DeleteTask(ctx context.Context, id string) error
}

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidInput = errors.New("invalid input")
)
