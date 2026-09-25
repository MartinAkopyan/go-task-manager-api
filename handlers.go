package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmptyTitle = errors.New("title field is required")
type TaskHandler struct {
	db *pgxpool.Pool
}

type UpdateTaskInput struct {
	Title *string `json:"title"`
	Done *bool `json:"done"`
}

func (t TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body Task
	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := CreateTask(r.Context(), t.db, body.Title)

	if errors.Is(err, ErrEmptyTitle) {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (t TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {

	tasks, err := GetTasks(r.Context(), t.db)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (t TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	strTaskID := chi.URLParam(r, "id")

	taskID, err := strconv.Atoi(strTaskID)

	if err != nil {
		http.Error(w, "Invalid task ID: must be an integer", http.StatusBadRequest)
		return
	}

	task, err := GetTaskByID(r.Context(), t.db, taskID)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (t TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	strTaskID := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(strTaskID)

	if  err != nil {
		http.Error(w, "Invalid task id: must be integer", http.StatusBadRequest)
		return
	}

	var input UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	task, err := UpdateTaskByID(r.Context(), t.db, taskID, input)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func CreateTask(ctx context.Context, db *pgxpool.Pool, title string) (int, error) {
	var id int

	if len(strings.TrimSpace(title)) <= 0 {
		return 0, ErrEmptyTitle
	}

	row := db.QueryRow(ctx, "INSERT INTO tasks(title, done) VALUES($1, $2) RETURNING id", title, false)

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func GetTasks(ctx context.Context, db *pgxpool.Pool) ([]Task, error) {
	tasks := []Task{}

	rows, err := db.Query(ctx, "SELECT id, title, done FROM tasks;")

	
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var t Task

		if err := rows.Scan(&t.ID, &t.Title, &t.Done); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}


func GetTaskByID(ctx context.Context, db *pgxpool.Pool, id int) (*Task, error) {
	var t Task

	row := db.QueryRow(ctx, "SELECT id, title, done FROM tasks WHERE id = $1", id)

	if err := row.Scan(&t.ID, &t.Title, &t.Done); err != nil {
		return nil, err
	}

	return &t, nil
}

func UpdateTaskByID(ctx context.Context, db *pgxpool.Pool, id int, input UpdateTaskInput) (*Task, error) {
	var t Task

	row := db.QueryRow(ctx, `
		UPDATE tasks 
		SET title = COALESCE($1, title),
			done = COALESCE($2, done)
		WHERE id = $3
		RETURNING id, title, done`, input.Title, input.Done, id)

	if err := row.Scan(&t.ID, &t.Title, &t.Done); err != nil {
		return nil, err
	}

	return &t, nil
}

// claude --resume 8e6cf020-253b-486b-8644-60c29d93ca82