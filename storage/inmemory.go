package storage

import (
	"context"
	"github.com/google/uuid"
	"sync"
	"tasks/models"
)

type InMemoryStorage struct {
	tasks map[string]models.Task
	mu    sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		tasks: make(map[string]models.Task),
	}
}

func (s *InMemoryStorage) CreateTask(ctx context.Context, task models.Task) (models.Task, error) {
	select {
	case <-ctx.Done():
		return models.Task{}, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if task.Title == "" {
		return models.Task{}, ErrInvalidInput
	}

	task.ID = uuid.New().String()
	task.Completed = false
	s.tasks[task.ID] = task
	return task, nil
}

func (s *InMemoryStorage) GetTaskByID(ctx context.Context, id string) (models.Task, error) {
	select {
	case <-ctx.Done():
		return models.Task{}, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	task, found := s.tasks[id]
	if !found {
		return models.Task{}, ErrTaskNotFound
	}
	return task, nil
}

func (s *InMemoryStorage) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tasksList := make([]models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasksList = append(tasksList, task)
	}
	return tasksList, nil
}

func (s *InMemoryStorage) UpdateTask(ctx context.Context, task models.Task) (models.Task, error) {
	select {
	case <-ctx.Done():
		return models.Task{}, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existingTask, found := s.tasks[task.ID]
	if !found {
		return models.Task{}, ErrTaskNotFound
	}

	if task.Title != "" {
		existingTask.Title = task.Title
	}

	existingTask.Completed = task.Completed

	s.tasks[task.ID] = existingTask
	return existingTask, nil
}

func (s *InMemoryStorage) DeleteTask(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, found := s.tasks[id]
	if !found {
		return ErrTaskNotFound
	}
	delete(s.tasks, id)
	return nil
}
