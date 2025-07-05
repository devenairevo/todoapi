# Todo API

A simple RESTful API for managing todo tasks, built with Go.

## Features

- Create, read, update, and delete todo tasks
- Docker compose file for Postgres setup
- In-memory storage for easy setup and testing
- Storage variations: preferable data saving (Postgres or In-memory)
- Modular code structure for easy extension

## Project Structure

- `main.go` – Application entry point and server setup
- `db/` – migrations, interface
- `middleware/` – middlewares
- `handlers/` – HTTP handlers
- `models/` – Data models (e.g., Task)
- `storage/` – Storage abstraction and in-memory implementation
- `test/requests` – HTTP requests via Jetbrains HTTP tool

## Getting Started

1. **Install dependencies:**
   ```sh
   go mod tidy
   ```

2. **Run the server:**
   ```sh
   go run main.go
   ```

3. **API Endpoints:**
   - `GET /v1/tasks` – List all tasks
   - `POST /v1/tasks` – Create a new task
   - `GET /v1/tasks/{id}` – Get a task by ID
   - `PUT /v1/tasks/{id}` – Update a task
   - `DELETE /v1/tasks/{id}` – Delete a task
