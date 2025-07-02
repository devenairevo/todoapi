package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tasks/api"
	"tasks/storage"
	"time"
)

func main() {
	taskStorage := storage.NewInMemoryStorage()

	apiConfig := api.NewAPIConfig(taskStorage)

	mux := http.NewServeMux()

	publicGetTasksHandler := api.LoggingMiddleware(http.HandlerFunc(apiConfig.GetTasksHandler))
	authenticatedCreateTaskHandler := api.LoggingMiddleware(api.AuthMiddleware(http.HandlerFunc(apiConfig.CreateTaskHandler)))

	mux.Handle("/tasks", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			publicGetTasksHandler.ServeHTTP(w, r)
		case http.MethodPost:
			authenticatedCreateTaskHandler.ServeHTTP(w, r)
		default:
			api.HandleError(w, errors.New("Method Not Allowed"))
		}
	}))

	publicHandleTaskByIDGet := api.LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiConfig.HandleTaskByID(w, r)
	}))

	authenticatedHandleTaskByID := api.LoggingMiddleware(api.AuthMiddleware(http.HandlerFunc(apiConfig.HandleTaskByID)))

	mux.Handle("/tasks/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			publicHandleTaskByIDGet.ServeHTTP(w, r)
		} else {
			authenticatedHandleTaskByID.ServeHTTP(w, r)
		}
	}))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

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
