package service

import (
	"sync"

	"calculator/pkg/task"

	"github.com/google/uuid"
)

type TaskService struct {
	tasks   []task.Task
	futures map[string]*task.Future
	mu      sync.Mutex
}

func NewTaskService() *TaskService {
	return &TaskService{
		futures: make(map[string]*task.Future),
	}
}

func (s *TaskService) CreateTask(arg1, arg2 float64, op string, opTime int) task.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskk := task.Task{
		ID:            uuid.New().String(),
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     op,
		OperationTime: opTime,
	}

	future := task.NewFuture()
	s.futures[taskk.ID] = future
	s.tasks = append(s.tasks, taskk)

	return taskk
}

func (s *TaskService) GetNextTask() (task.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.tasks) == 0 {
		return task.Task{}, false
	}

	task := s.tasks[0]
	s.tasks = s.tasks[1:]
	return task, true
}

func (s *TaskService) SetResult(id string, result float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if future, exists := s.futures[id]; exists {
		future.SetResult(result)
		delete(s.futures, id)
		return true
	}
	return false
}
