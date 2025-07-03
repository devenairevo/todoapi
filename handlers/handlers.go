package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/devenairevo/todoapi/models"
	"github.com/devenairevo/todoapi/storage"
	"log"
	"net/http"
)

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

type TaskStorage struct {
	storage.Task
}

func NewTaskStorage(s storage.Task) *TaskStorage {
	return &TaskStorage{Task: s}
}

func (ts *TaskStorage) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HandleError(w, errors.New("method Not Allowed"))
		return
	}

	var newTask models.Task
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		HandleError(w, errors.Join(storage.ErrInvalidInput, err))
		return
	}

	createdTask, err := ts.Create(r.Context(), newTask)
	if err != nil {
		HandleError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusCreated, createdTask)
	log.Printf("Created task: %+v\n", createdTask)
}

func (ts *TaskStorage) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HandleError(w, errors.New("method Not Allowed"))
		return
	}

	tasksList, err := ts.GetAll(r.Context())
	if err != nil {
		HandleError(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, tasksList)
}

func (ts *TaskStorage) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		HandleError(w, errors.Join(storage.ErrInvalidInput, errors.New("task ID is required")))
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, err := ts.GetByID(r.Context(), id) // Передаем контекст
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

		updatedTask, err := ts.Update(r.Context(), updatedTaskData)
		if err != nil {
			HandleError(w, err)
			return
		}
		writeJSONResponse(w, http.StatusOK, updatedTask)
		log.Printf("Updated task %s: %+v\n", id, updatedTask)

	case http.MethodDelete:
		err := ts.Delete(r.Context(), id)
		if err != nil {
			HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Printf("Deleted task: %s\n", id)

	default:
		HandleError(w, errors.New("method Not Allowed"))
	}
}
