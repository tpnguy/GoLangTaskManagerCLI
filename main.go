package main

import (
	// "bytes"
	"encoding/json"
	"sync"
	// "fmt"
	"net/http"
	"os"
	"strings"
	// "log"
	"strconv"
)

type Task struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type App struct {
	Tasks []Task
	Mu sync.Mutex
}


func loadTasks() []Task {

	data, err := os.ReadFile("tasks.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}
		}
		panic(err)
	}

	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		panic(err)
	}

	return tasks
}

func saveTasks(tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		return err	
	}
	err = os.WriteFile("tasks.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) getTasks(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")

	a.Mu.Lock()
	defer a.Mu.Unlock()

	enc := json.NewEncoder(w)

	enc.SetIndent("", " ")
	if err := enc.Encode(a.Tasks); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}	
}

func (a *App) postTasks(w http.ResponseWriter, r *http.Request){
	
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	a.Mu.Lock()
	defer a.Mu.Unlock()

	task.ID = len(a.Tasks) + 1
	a.Tasks = append(a.Tasks, task)
	saveTasks(a.Tasks)

	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := enc.Encode(task); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (a *App) deleteTask(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	index, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if index < 0 || index >= len(a.Tasks){
		http.Error(w, "task index out of range", http.StatusNotFound)
		return
	}

	a.Mu.Lock()
	defer a.Mu.Unlock()

	a.Tasks = append(a.Tasks[:index], a.Tasks[index+1:]...)
	if err := saveTasks(a.Tasks); err != nil {
		http.Error(w, "failed to save tasks", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message":"deleted",
		"index": index,
	})
}

func (a *App) tasksHandler(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		a.getTasks(w, r)
	case http.MethodPost:
		a.postTasks(w,r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	app := &App {
		Tasks: loadTasks(),
	}
	http.HandleFunc("/tasks", app.tasksHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
