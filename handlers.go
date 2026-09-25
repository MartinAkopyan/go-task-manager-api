package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmptyTitle = errors.New("title field is required")
type TaskHandler struct {
	db *pgxpool.Pool
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

// claude --resume 8e6cf020-253b-486b-8644-60c29d93ca82