package handler

import (
	"errors"
	"net/http"

	"calculator/internal/service"
	"calculator/pkg/calculator"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(s *service.TaskService) *TaskHandler {
	return &TaskHandler{service: s}
}

type TaskRequest struct {
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	var opTime int
	switch req.Operation {
	case "+":
		opTime = 1000 // default add time
	case "-":
		opTime = 1000 // default subtract time
	case "*":
		opTime = 1000 // default multiply time
	case "/":
		opTime = 1000 // default divide time
	default:
		render.Render(w, r, ErrBadRequest(errors.New("invalid operation")))
		return
	}

	task := h.service.CreateTask(req.Arg1, req.Arg2, req.Operation, opTime)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, exists := h.service.GetNextTask()
	if !exists {
		render.Render(w, r, ErrNotFound(errors.New("no tasks available")))
		return
	}
	render.JSON(w, r, task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var result struct {
		Result float64 `json:"result"`
	}
	if err := render.DecodeJSON(r.Body, &result); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	if !h.service.SetResult(id, result.Result) {
		render.Render(w, r, ErrNotFound(errors.New("task not found")))
		return
	}

	render.NoContent(w, r)
}

type CalculationRequest struct {
	Expression string `json:"expression"`
}

type CalculationResponse struct {
	Result float64 `json:"result"`
}

func (h *TaskHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req CalculationRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	calc := calculator.New(h.service)
	result, err := calc.Calculate(req.Expression)
	if err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	render.JSON(w, r, CalculationResponse{Result: result})
}

func (h *TaskHandler) Index(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{
				"error": "Internal server error",
			})
		}
	}()
}
