package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	// "log"
	"strconv"
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

func saveTasks(tasks []Task) {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil{
		panic(err)
	}
	err = os.WriteFile("tasks.json", data, 0644)
	if err != nil{
		panic(err)
	}
}

func updateTask(index int, tasks []Task){
	if index >= 0 && index < len(tasks){
		tasks[index].Done = true
		fmt.Println("Task Updated.")
		saveTasks(tasks)
	} else {
		fmt.Println("Task index out of range.")
	}
}

func deleteIndex(index int, tasks []Task) []Task {
	if index >= 0 && index < len(tasks){
		tasks = append(tasks[:index], tasks[index+1:]...)
		fmt.Println("Task deleted.")
		saveTasks(tasks)
		return tasks
	} else{
		fmt.Println("Task index out of range.")
		return tasks
	}
}

func main() {

	tasks := loadTasks()

	// fmt.Println(tasks)
	if len(os.Args) <= 1 {
		fmt.Println("Usage: main.go list | main.go add | main.go done")
		return
	}

	command := os.Args[1]

	switch command {
	case "list":
		for i, v := range tasks {
			status := "[ ]"
			if v.Done {
				status = "[X]"
			}
			fmt.Println(i, status, v.Title)
		}
	case "add":
		if len(os.Args) > 2 {
			var newTitle = strings.Join(os.Args[2:], " ")

			newTask := Task{
				Title: newTitle,
				Done:  false,
			}
			tasks = append(tasks, newTask)
			fmt.Printf("Task: %q Added", newTask.Title)
			saveTasks(tasks)
		} else {
			fmt.Printf("Need to have another argument.")
		}
	case "done":
		if len(os.Args) > 2 {
			index, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println("Error occured converting integer.")
				return
			}
			updateTask(index, tasks)
		}
	case "delete":
		if len(os.Args) > 2 {
			index, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println("Error occured converting integer.")
				return
			}
			tasks = deleteIndex(index, tasks)
		}
	default:
		fmt.Println("None of the options matched.")
	}
}
