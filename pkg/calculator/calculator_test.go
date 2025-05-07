package calculator

import (
	"calculator/internal/service"
	"testing"
)

func TestCalculator(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected float64
		wantErr  bool
	}{
		{"simple addition", "2 + 2", 4, false},
		{"simple subtraction", "5 - 3", 2, false},
		{"simple multiplication", "3 * 4", 12, false},
		{"simple division", "8 / 2", 4, false},
		{"complex expression", "3 + 4 * 2 / (1 - 5)", 1, false},
		{"division by zero", "5 / 0", 0, true},
		{"invalid character", "2 # 3", 0, true},
		{"invalid brackets", "(2 + 3", 0, true},
	}

	calc := New(service.NewTaskService())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Calculate(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("Calculate() = %v, want %v", result, tt.expected)
			}
		})
	}
}
