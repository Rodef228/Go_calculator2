package agent

import (
	"log"
	"strconv"
	"time"

	"calculator/pkg/config"
	"calculator/pkg/models"
)

func worker(cfg config.Config) {
	for task := range tasksCh {
		log.Printf("worker got expression with id %v", task.ID)
		result, err := calculate(task.Arg1, task.Arg2, task.Type, cfg)

		res := &models.Result{ID: task.ID, Result: result, Error: err}
		resultsCh <- res
		log.Printf("worker sent result %v with id %v", result, task.ID)
	}
}

func calculate(a, b string, operator string, cfg config.Config) (float64, string) {
	a_float, _ := strconv.ParseFloat(a, 64)
	b_float, _ := strconv.ParseFloat(b, 64)

	switch operator {
	case "*":
		time.Sleep(time.Duration(cfg.MultiplyTimeMs))
		return a_float * b_float, ""
	case "/":
		time.Sleep(time.Duration(cfg.DivideTimeMs))
		if b_float == 0 {
			return 0, "division by zero"
		}
		return a_float / b_float, ""
	case "+":
		time.Sleep(time.Duration(cfg.AddTimeMs))
		return a_float + b_float, ""
	case "-":
		time.Sleep(time.Duration(cfg.DivideTimeMs))
		return a_float - b_float, ""
	default:
		return 0, ""
	}
}
