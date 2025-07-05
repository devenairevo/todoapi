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
	"time"
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

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return models.Task{}, ctx.Err()
	default:
	}

	query := `
		SELECT id, title, completed 
		FROM tasks 
		WHERE id = $1
		`
	row := pg.Pool.QueryRow(ctx, query, id)

	var task models.Task
	err := row.Scan(&task.ID, &task.Title, &task.Completed)
	if err != nil {
		return models.Task{}, fmt.Errorf("couldn't retreive the task by ID")
	}

	return task, nil
}

func (pg *InPostgresql) GetAll(ctx context.Context) ([]models.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return []models.Task{}, ctx.Err()
	default:
	}

	query := `
		SELECT id, title, completed 
		FROM tasks
		`
	rows, err := pg.Pool.Query(ctx, query)
	if err != nil {
		fmt.Errorf("couldn't retreive tasks")
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Completed)
		if err != nil {
			return nil, fmt.Errorf("couldn't retreive the specific task in a rows %s", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (pg *InPostgresql) Update(ctx context.Context, task models.Task) (models.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return models.Task{}, ctx.Err()
	default:
	}

	updateQuery := `
		UPDATE tasks
		SET title = $2,
			completed = $3
		WHERE id = $1
		RETURNING id, title, completed;
	`

	row := pg.Pool.QueryRow(ctx, updateQuery, task.ID, task.Title, task.Completed)

	var updatedTask models.Task
	err := row.Scan(&updatedTask.ID, &updatedTask.Title, &updatedTask.Completed)
	if err != nil {
		return models.Task{}, fmt.Errorf("couldn't retreive the specific task %s", err)
	}

	return task, nil
}

func (pg *InPostgresql) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	removeQuery := `
		DELETE FROM tasks
		WHERE id = $1
	`

	tag, err := pg.Pool.Exec(ctx, removeQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no task found with id %s", id)
	}

	return nil
}
