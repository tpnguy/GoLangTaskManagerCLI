package main

import (
	"encoding/json"
	"errors"

	"net/http"
	"strings"

	"strconv"

	"database/sql"

	_ "modernc.org/sqlite"

	"log"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type UpdateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

type App struct {
	DB     *sql.DB
}

func parseTaskID(r *http.Request) (int, error) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	return strconv.Atoi(path)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(append(data, '\n'))
	return err
}


func (a *App) getTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query("SELECT id, title, done FROM tasks")
	if err != nil {
		http.Error(w, "failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Done); err != nil {
			http.Error(w, "failed to scan task", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusOK, tasks); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (a *App) getTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)

	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var task Task
	row := a.DB.QueryRow(`SELECT id, title, done FROM tasks WHERE id = ?`, id)
	if err := row.Scan(&task.ID, &task.Title, &task.Done); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch task", http.StatusInternalServerError)
		return
	}
	if err := writeJSON(w, http.StatusOK, task); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (a *App) postTasks(w http.ResponseWriter, r *http.Request) {
	var task Task
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&task); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	result, err := a.DB.Exec(
		"INSERT INTO tasks (title, done) VALUES (?, ?)",
		task.Title,
		task.Done,
	)
	if err != nil {
		http.Error(w, "failed to insert task", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "failed to retrieve id", http.StatusInternalServerError)
		return
	}

	task.ID = int(id)
	task.Done = false

	if err := writeJSON(w, http.StatusCreated, task); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (a *App) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	result, err := a.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		http.Error(w, "delete query failed", http.StatusInternalServerError)
		return
	}

	n, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "failed to read delete result", http.StatusInternalServerError)
		return
	}
	if n == 0 {
		http.Error(w, "no row found", http.StatusNotFound)
		return
	}

	if err := writeJSON(w, http.StatusOK, map[string]any{
		"message": "deleted",
		"id":      id,
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (a *App) updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}
	var update UpdateTaskRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&update); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if update.Title == nil && update.Done == nil {
		http.Error(w, "no fields provided for update", http.StatusBadRequest)
		return
	}

	if update.Title != nil && *update.Title == "" {
		http.Error(w, "title cannot be empty", http.StatusBadRequest)
		return
	}

	result, err := a.DB.Exec(
		"UPDATE tasks set title = COALESCE(?, title), done = COALESCE(?, done) WHERE id = ?", update.Title, update.Done, id,
	)	

	if err != nil {
		http.Error(w, "unable to update database", http.StatusInternalServerError)
		return
	}

	n, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "failed to read update result", http.StatusInternalServerError)
		return
	}

	if n == 0 {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	var task Task
	row := a.DB.QueryRow("SELECT id, title, done FROM tasks WHERE id = ?", id)
	if err := row.Scan(&task.ID, &task.Title, &task.Done); err != nil {
		http.Error(w, "failed to fetch updated task", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusOK, task); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (a *App) tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getTasks(w, r)
	case http.MethodPost:
		a.postTasks(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) tasksByIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getTaskByID(w, r)
	case http.MethodDelete:
		a.deleteTask(w, r)
	case http.MethodPatch:
		a.updateTask(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {

	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT 0
	);
	`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	app := &App{
		DB: db,
	}

	http.HandleFunc("/tasks", app.tasksHandler)
	http.HandleFunc("/tasks/", app.tasksByIDHandler)
	log.Println("Server listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
