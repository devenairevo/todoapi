package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/devenairevo/todoapi/handlers"
	"github.com/devenairevo/todoapi/middleware"
	"github.com/devenairevo/todoapi/storage"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var taskHandler *handlers.TaskStorage

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	storageType := os.Getenv("STORAGE")

	// Storage type check
	switch storageType {
	case "inpostgres":
		dbStorage, err := storage.NewInPostgresStorage()
		if err != nil {
			fmt.Errorf("connection error %s", err)
		}
		defer dbStorage.Pool.Close()

		taskHandler = handlers.NewTaskStorage(dbStorage)
	case "inmemory":
		inMemoryStorage := storage.NewInMemoryStorage()
		taskHandler = handlers.NewTaskStorage(inMemoryStorage)
	default:
		log.Fatalf("Unknown storage type: %s", storageType)
	}

	router := http.NewServeMux()

	v1 := http.NewServeMux()
	v1.Handle("/", http.StripPrefix("/v1", router))

	server := &http.Server{
		Addr: ":8080",
		Handler: middleware.LoggingMiddleware(
			middleware.AuthMiddleware(http.StripPrefix("/v1", router)),
		),
	}

	// Requests
	router.HandleFunc("POST /tasks", taskHandler.CreateTask)
	router.HandleFunc("GET /tasks", taskHandler.GetTasks)
	router.HandleFunc("GET /tasks/{id}", taskHandler.GetTaskByID)
	router.HandleFunc("PUT /tasks/{id}", taskHandler.GetTaskByID)
	router.HandleFunc("DELETE /tasks/{id}", taskHandler.GetTaskByID)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server starting on port 8080...")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited gracefully.")
}
