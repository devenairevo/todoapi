package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"tasks/models"
	"tasks/storage"
)

type Config struct {
	Storage storage.TaskStorage
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HandleError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, storage.ErrTaskNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, storage.ErrInvalidInput):
		statusCode = http.StatusBadRequest
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		statusCode = http.StatusRequestTimeout
		log.Printf("Request canceled/timeout: %v", err)
	default:
		log.Printf("Unhandled server error: %v", err)
	}
	writeJSONResponse(w, statusCode, map[string]string{"error": err.Error()})
}

func NewAPIConfig(s storage.TaskStorage) *Config {
	return &Config{Storage: s}
}

func (cfg *Config) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HandleError(w, errors.New("method Not Allowed"))
		return
	}

	var newTask models.Task
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		HandleError(w, errors.Join(storage.ErrInvalidInput, err))
		return
	}

	createdTask, err := cfg.Storage.CreateTask(r.Context(), newTask)
	if err != nil {
		HandleError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusCreated, createdTask)
	log.Printf("Created task: %+v\n", createdTask)
}

func (cfg *Config) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HandleError(w, errors.New("method Not Allowed"))
		return
	}

	tasksList, err := cfg.Storage.GetAllTasks(r.Context())
	if err != nil {
		HandleError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, tasksList)
}

func (cfg *Config) HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		HandleError(w, errors.Join(storage.ErrInvalidInput, errors.New("task ID is required")))
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, err := cfg.Storage.GetTaskByID(r.Context(), id) // Передаем контекст
		if err != nil {
			HandleError(w, err)
			return
		}
		writeJSONResponse(w, http.StatusOK, task)

	case http.MethodPut:
		var updatedTaskData models.Task
		if err := json.NewDecoder(r.Body).Decode(&updatedTaskData); err != nil {
			HandleError(w, errors.Join(storage.ErrInvalidInput, err))
			return
		}
		updatedTaskData.ID = id

		updatedTask, err := cfg.Storage.UpdateTask(r.Context(), updatedTaskData)
		if err != nil {
			HandleError(w, err)
			return
		}
		writeJSONResponse(w, http.StatusOK, updatedTask)
		log.Printf("Updated task %s: %+v\n", id, updatedTask)

	case http.MethodDelete:
		err := cfg.Storage.DeleteTask(r.Context(), id)
		if err != nil {
			HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		log.Printf("Deleted task: %s\n", id)

	default:
		HandleError(w, errors.New("method Not Allowed"))
	}
}
