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

type UpdateTaskRequest struct {
	Title *string `json:"title"`
	Done *bool `json:"done"`
}

type App struct {
	Tasks []Task
	NextID int
	Mu sync.RWMutex
}

func findTaskIndexById(tasks []Task, index int) int {
	for i := range tasks {
		if tasks[i].ID == index{
			return i
		}
	}
	return -1
}

func parseTaskID(r *http.Request) (int, error) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	return strconv.Atoi(path)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(append(data, '\n'))
	return err
}

func nextTaskID(tasks []Task) int {
	maxID := 0
	for _, t := range tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
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
	a.Mu.Lock()
	defer a.Mu.Unlock()

	if err := writeJSON(w, http.StatusOK, a.Tasks); err != nil {
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

	a.Mu.RLock()
	defer a.Mu.RUnlock()
	
	foundIndex := findTaskIndexById(a.Tasks, id)
	if foundIndex != -1 {
		if err := writeJSON(w, http.StatusOK, a.Tasks[foundIndex]); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Task not found.", http.StatusNotFound)	
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

	task.ID = a.NextID
	a.NextID++
	a.Tasks = append(a.Tasks, task)
	if err := saveTasks(a.Tasks); err != nil {
		http.Error(w, "unable to save task", http.StatusInternalServerError)
		return
	}
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

	a.Mu.Lock()
	defer a.Mu.Unlock()

	foundIndex := findTaskIndexById(a.Tasks, id)

	if foundIndex == -1 {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	a.Tasks = append(a.Tasks[:foundIndex], a.Tasks[foundIndex+1:]...)
	if err := saveTasks(a.Tasks); err != nil {
		http.Error(w, "failed to save tasks", http.StatusInternalServerError)
		return
	}
	if err := writeJSON(w, http.StatusOK, map[string]any{
		"message":"deleted",
		"id": id,
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
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
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

	a.Mu.Lock()
	defer a.Mu.Unlock()

	foundIndex := findTaskIndexById(a.Tasks, id)
	if foundIndex != -1 {
		if update.Title != nil {
			a.Tasks[foundIndex].Title = *update.Title
		}
		if update.Done != nil {
			a.Tasks[foundIndex].Done = *update.Done
		}
		if err := saveTasks(a.Tasks); err != nil {
			http.Error(w, "Unable to save task.", http.StatusInternalServerError)
			return
		}
		if err := writeJSON(w, http.StatusOK, a.Tasks[foundIndex]); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
		return
	} else{
		http.Error(w, "Task not found.", http.StatusNotFound)
	}

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

func (a *App) tasksByIDHandler(w http.ResponseWriter, r *http.Request){
	switch r.Method{
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
	tasks := loadTasks()
	app := &App {
		Tasks: tasks,
		NextID: nextTaskID(tasks),
	}
	http.HandleFunc("/tasks", app.tasksHandler)
	http.HandleFunc("/tasks/", app.tasksByIDHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
