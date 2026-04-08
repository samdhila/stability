package store

import "stability-test-task-api/models"

var Tasks = []models.Task{
	{ID: 1, Title: "Learn Go", Done: false},
	{ID: 2, Title: "Build API", Done: false},
}

func GetAllTasks() []models.Task {
	return Tasks
}

func GetTaskByID(id int) *models.Task {
	for _, t := range Tasks {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

var nextID = 3

func AddTask(task models.Task) models.Task {
	task.ID = nextID
	nextID++
	Tasks = append(Tasks, task)
	return task
}

func DeleteTask(id int) bool {
	for i, t := range Tasks {
		if t.ID == id {
			Tasks = append(Tasks[:i], Tasks[i+1:]...)
			return true
		}
	}
	return false
}

func UpdateTask(id int, updatedTask models.Task) *models.Task {
	for i, t := range Tasks {
		if t.ID == id {
			updatedTask.ID = id
			Tasks[i] = updatedTask
			return &Tasks[i]
		}
	}
	return nil
}
