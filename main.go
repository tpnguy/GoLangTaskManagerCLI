package main

import "fmt"
import "os"

type Task struct {
	Title string
	Done bool
}


func main(){
	task1 := Task{
		Title: "This is task 1",
		Done:  false,
	}
	task2 := Task{
		Title: "This is task 2",
		Done:  true,
	}
	task3 := Task{
		Title: "This is task 3",
		Done:  false,
	}

	tasks := []Task{task1, task2, task3}


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
			newTask := Task{
				Title: os.Args[2],
				Done:  false,
			}
			tasks = append(tasks, newTask)
			fmt.Printf("Task: \"%s\" Added", newTask.Title)
		} else {
			fmt.Printf("Need to have another argument.")
		}
	default:
		fmt.Println("None of the options matched.")
	}
}