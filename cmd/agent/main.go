package main

import (
	"bytes"
	"calculator/pkg/config"
	"calculator/pkg/logger"
	"calculator/pkg/task"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Agent struct {
	config config.Config
	log    logger.Logger
	client *http.Client
}

func New(cfg config.Config, log logger.Logger) *Agent {
	return &Agent{
		config: cfg,
		log:    log,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *Agent) Run() {
	sem := make(chan struct{}, a.config.AgentComputingPower)

	for {
		task, err := a.fetchTask()
		if err != nil {
			a.log.Error("failed to fetch task", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if task != nil {
			sem <- struct{}{}
			go a.ProcessTask(task, sem)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (a *Agent) fetchTask() (*task.Task, error) {
	fmt.Println(config.Configuration.OrchestratorURL + "/internal/task")
	resp, err := a.client.Get(config.Configuration.OrchestratorURL + "/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var task task.Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (a *Agent) ProcessTask(task *task.Task, sem chan struct{}) {
	defer func() { <-sem }()

	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	var result float64
	switch task.Operation {
	case "+":
		result = task.Arg1 + task.Arg2
	case "-":
		result = task.Arg1 - task.Arg2
	case "*":
		result = task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 == 0 {
			a.log.Error("division by zero", "taskID", task.ID)
			return
		}
		result = task.Arg1 / task.Arg2
	}

	if err := a.sendResult(task.ID, result); err != nil {
		a.log.Error("failed to send result", "taskID", task.ID, "error", err)
	}
}

func (a *Agent) sendResult(id string, result float64) error {
	data := struct {
		Result float64 `json:"result"`
	}{Result: result}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := a.client.Post(
		config.Configuration.OrchestratorURL+"/internal/task/"+id,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("unexpected status code")
	}

	return nil
}

func main() {
	config.Configuration.OrchestratorURL = "http://orchestrator:8080"
	config.LoadConfig()
	log := logger.New("agent")

	agent := New(config.Configuration, log)
	agent.Run()
}
