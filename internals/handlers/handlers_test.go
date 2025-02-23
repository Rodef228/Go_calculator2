package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"http_calculator/internals/models"
)

func TestCalculateHandler(t *testing.T) {
	reqBody := map[string]string{"expression": "2+2"}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CalculateHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if _, ok := response["id"]; !ok {
		t.Errorf("handler did not return an id")
	}
}

func TestExpressionsHandler(t *testing.T) {
	expressions["testID"] = models.Expression{ID: "testID", Status: "done", Result: 4}

	req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ExpressionsHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]map[string]models.Expression
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if _, ok := response["expressions"]["testID"]; !ok {
		t.Errorf("handler did not return the test expression")
	}
}

func TestExpressionHandler(t *testing.T) {
	expressions["testID"] = models.Expression{ID: "testID", Status: "done", Result: 4}

	req, err := http.NewRequest("GET", "/api/v1/expressions/testID", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ExpressionHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]models.Expression
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response["expression"].ID != "testID" {
		t.Errorf("handler returned wrong expression: got %v want %v", response["expression"].ID, "testID")
	}
}

func TestTaskHandler(t *testing.T) {
	tasks["taskID"] = models.Task{ID: "taskID", Arg1: 2, Arg2: 2, Operation: "+", OperationTime: 1000}

	req, err := http.NewRequest("GET", "/internal/task", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(TaskHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]models.Task
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response["task"].ID != "taskID" {
		t.Errorf("handler returned wrong task: got %v want %v", response["task"].ID, "taskID")
	}
}

func TestTaskResultHandler(t *testing.T) {
	expressions["testID"] = models.Expression{ID: "testID", Status: "pending", Result: 0}

	reqBody := map[string]interface{}{"id": "testID", "result": 4.0}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/internal/task", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(TaskResultHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if expressions["testID"].Result != 4.0 {
		t.Errorf("handler did not update the expression result: got %v want %v", expressions["testID"].Result, 4.0)
	}
}
