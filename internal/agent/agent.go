package agent

import (
	"log"

	"calculator/pkg/config"
	"calculator/pkg/models"
)

type Agent struct {
	config config.Config
}

type Task struct {
	ID   int
	Arg1 string
	Arg2 string
	Type string
}

var (
	resultsCh = make(chan *models.Result)
	tasksCh   = make(chan *Task)
)

func New(cfg config.Config) *Agent {
	// передаем конфиг с переменными средами в агента
	return &Agent{config: cfg}
}

func (a *Agent) Run() {
	tasksCh = make(chan *Task, a.config.AgentComputingPower)
	resultsCh = make(chan *models.Result, a.config.AgentComputingPower)
	go a.Connect()

	for i := range a.config.AgentComputingPower {
		log.Printf("worker %d starting...", i+1)
		go worker(a.config)
	}

	select {} // бесконечное ожидание
}
