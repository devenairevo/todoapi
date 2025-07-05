package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"sync"
)

type Postgresql struct {
	Pool *pgxpool.Pool
}

var once sync.Once
var instance *Postgresql

func (p *Postgresql) Connect() *Postgresql {
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

		instance = &Postgresql{
			Pool: pool,
		}
	})

	return instance
}
