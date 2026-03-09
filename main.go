package main

import (
	// "bytes"
	"encoding/json"
	"sync"
	// "fmt"
	// "net/http"
	"os"
	// "strings"
	// "log"
	// "strconv"
)

type Task struct {
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

func saveTasks(tasks []Task) {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("tasks.json", data, 0644)
	if err != nil {
		panic(err)
	}
}



func main() {

	
}
