package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	// "log"
)

type Task struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func loadTasks() []Task {

	data, err := os.ReadFile("tasks.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}
		}
		panic(err)
	}

	// if len(bytes.TrimSpace(data)) == 0 {
	// 	return []Task{}
	// }

	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		panic(err)
	}

	return tasks
}

func saveTask(tasks []Task) {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("tasks.json", data, 0644)
	if err != nil{
		panic(err)
	}
}

func main() {

	tasks := loadTasks()

	// fmt.Println(tasks)
	if len(os.Args) <= 1 {
		fmt.Println("Usage: main.go list | main.go add")
		return
	}

	command := os.Args[1]

	switch command {
	case "list":
		for i, v := range tasks {
			fmt.Println(i, v.Title, v.Done)
		}
	case "add":
		if len(os.Args) > 2 {
			var newTitle = strings.Join(os.Args[2:], " ")

			newTask := Task{
				Title: newTitle,
				Done:  false,
			}
			tasks = append(tasks, newTask)
			fmt.Printf("Task: \"%s\" Added", newTask.Title)
			saveTask(tasks)
		} else {
			fmt.Printf("Need to have another argument.")
		}
	default:
		fmt.Println("None of the options matched.")
	}
}
