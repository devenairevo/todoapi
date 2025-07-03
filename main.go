package main

import (
	"context"
	"errors"
	"github.com/devenairevo/todoapi/handlers"
	"github.com/devenairevo/todoapi/middleware"
	"github.com/devenairevo/todoapi/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	inMemoryStorage := storage.NewInMemoryStorage()
	taskStorageHandler := handlers.NewTaskStorage(inMemoryStorage)
	router := http.NewServeMux()

	server := &http.Server{
		Addr: ":8080",
		Handler: middleware.LoggingMiddleware(
			middleware.AuthMiddleware(router),
		),
	}

	// Requests
	router.HandleFunc("POST /tasks", taskStorageHandler.CreateTask)
	router.HandleFunc("GET /tasks", taskStorageHandler.GetTasks)
	router.HandleFunc("GET /tasks/{id}", taskStorageHandler.GetTaskByID)
	router.HandleFunc("PUT /tasks/{id}", taskStorageHandler.GetTaskByID)
	router.HandleFunc("DELETE /tasks/{id}", taskStorageHandler.GetTaskByID)

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
