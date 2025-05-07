package main

import (
	"calculator/pkg/config"
	"calculator/pkg/logger"
	"calculator/pkg/task"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestAgent(t *testing.T) {
	log := logger.NewMockLogger()
	tests := []struct {
		name          string
		config        config.Config
		taskResponse  *task.Task
		expectError   bool
		expectedCalls int
	}{
		{
			name: "successful task processing",
			config: config.Config{
				AgentComputingPower: 2,
			},
			taskResponse: &task.Task{
				ID:            "test1",
				Arg1:          10,
				Arg2:          5,
				Operation:     "+",
				OperationTime: 100,
			},
			expectedCalls: 1,
		},
		{
			name: "division by zero",
			config: config.Config{
				AgentComputingPower: 1,
			},
			taskResponse: &task.Task{
				ID:            "test2",
				Arg1:          10,
				Arg2:          0,
				Operation:     "/",
				OperationTime: 100,
			},
			expectedCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			var mu sync.Mutex
			var callCount int

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()

				switch r.URL.Path {
				case "/internal/task":
					if tt.taskResponse == nil {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					json.NewEncoder(w).Encode(tt.taskResponse)
				case "/internal/task/" + tt.taskResponse.ID:
					callCount++
					w.WriteHeader(http.StatusNoContent)
				}
			}))
			defer server.Close()

			// Create agent with test config
			a := New(tt.config, log)
			a.client = server.Client() // Override client to use mock server

			// Replace URL with mock server URL
			oldTransport := a.client.Transport
			defer func() { a.client.Transport = oldTransport }()
			a.client.Transport = &rewriteTransport{
				Transport: oldTransport,
				URL:       server.URL,
			}

			// Run agent for short duration
			go a.Run()
			time.Sleep(300 * time.Millisecond) // Wait for processing

			// Verify results
			mu.Lock()
			if callCount != tt.expectedCalls {
				t.Errorf("expected %d calls to sendResult, got %d", tt.expectedCalls, callCount)
			}
			mu.Unlock()
		})
	}
}

// TestProcessTask directly tests the ProcessTask method
func TestProcessTask(t *testing.T) {
	log := logger.NewMockLogger()
	cfg := config.Config{AgentComputingPower: 1}

	// Setup mock server
	var resultSent float64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/task/test3" {
			var data struct {
				Result float64 `json:"result"`
			}
			json.NewDecoder(r.Body).Decode(&data)
			resultSent = data.Result
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	// Create agent
	a := New(cfg, log)
	a.client = server.Client()

	// Test cases
	tests := []struct {
		name     string
		task     *task.Task
		expected float64
	}{
		{
			name: "addition",
			task: &task.Task{
				ID:            "test3",
				Arg1:          8,
				Arg2:          4,
				Operation:     "+",
				OperationTime: 10,
			},
			expected: 12,
		},
		{
			name: "multiplication",
			task: &task.Task{
				ID:            "test4",
				Arg1:          5,
				Arg2:          6,
				Operation:     "*",
				OperationTime: 10,
			},
			expected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sem := make(chan struct{}, 1)
			a.ProcessTask(tt.task, sem)

			if resultSent != tt.expected {
				t.Errorf("expected result %f, got %f", tt.expected, resultSent)
			}
		})
	}
}

// TestFetchTask tests the fetchTask method
func TestFetchTask(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/task" {
			json.NewEncoder(w).Encode(&task.Task{
				ID:            "test5",
				Arg1:          15,
				Arg2:          3,
				Operation:     "-",
				OperationTime: 50,
			})
		}
	}))
	defer server.Close()

	// Create agent
	a := New(config.Config{}, logger.NewMockLogger())
	a.client = server.Client()

	task, err := a.fetchTask()
	if err != nil {
		t.Fatalf("fetchTask failed: %v", err)
	}

	if task.ID != "test5" {
		t.Errorf("expected task ID 'test5', got '%s'", task.ID)
	}
}

// rewriteTransport rewrites requests to use mock server
type rewriteTransport struct {
	Transport http.RoundTripper
	URL       string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.URL[len("http://"):]
	return t.Transport.RoundTrip(req)
}
