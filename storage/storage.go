package storage

import (
	"context"
	"errors"
	"github.com/devenairevo/todoapi/models"
)

type Tasker interface {
	Create(ctx context.Context, task models.Task) (models.Task, error)
	GetByID(ctx context.Context, id string) (models.Task, error)
	GetAll(ctx context.Context) ([]models.Task, error)
	Update(ctx context.Context, task models.Task) (models.Task, error)
	Delete(ctx context.Context, id string) error
}

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidInput = errors.New("invalid input")
)
