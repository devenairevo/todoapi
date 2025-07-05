package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/devenairevo/todoapi/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"sync"
)

type InPostgresql struct {
	Pool *pgxpool.Pool
}

var once sync.Once
var instance *InPostgresql

func (pg *InPostgresql) Connect() (*InPostgresql, error) {
	once.Do(func() {
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)

		pool, err := pgxpool.New(context.Background(), dbURL)
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to DB: %v", err))
		}

		instance = &InPostgresql{
			Pool: pool,
		}
	})

	return instance, nil
}

type InPostgresStorage struct {
	DB *sql.DB
	mu sync.RWMutex
}

func NewInPostgresStorage() (*InPostgresql, error) {
	pg := &InPostgresql{}
	conn, err := pg.Connect()
	if err != nil {
		return nil, fmt.Errorf("Server forced to shutdown: %v\n", err)
	}
	return conn, nil
}

func (pg *InPostgresql) Create(ctx context.Context, task models.Task) (models.Task, error) {
	select {
	case <-ctx.Done():
		return models.Task{}, ctx.Err()
	default:
	}

	task.ID = uuid.New().String()
	if task.Title == "" {
		return models.Task{}, ErrInvalidInput
	}

	query := `
		INSERT INTO tasks (id, title, completed)
		VALUES ($1, $2, $3)
		RETURNING id, title, completed`
	row := pg.Pool.QueryRow(ctx, query, task.ID, task.Title, task.Completed)

	var createdTask models.Task
	err := row.Scan(&createdTask.ID, &createdTask.Title, &createdTask.Completed)
	if err != nil {
		return models.Task{}, fmt.Errorf("couldn't retreive the task")
	}

	return createdTask, nil

}

func (pg *InPostgresql) GetByID(ctx context.Context, id string) (models.Task, error) {
	return models.Task{}, nil
}

func (pg *InPostgresql) GetAll(ctx context.Context) ([]models.Task, error) {
	return []models.Task{}, nil
}

func (pg *InPostgresql) Update(ctx context.Context, task models.Task) (models.Task, error) {
	return models.Task{}, nil
}

func (pg *InPostgresql) Delete(ctx context.Context, id string) error {
	return nil
}
