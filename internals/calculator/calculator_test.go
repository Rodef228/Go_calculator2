package calculator

import (
	"errors"
	"testing"
)

func TestCalc(t *testing.T) {
	tests := []struct {
		expression string
		expected   float64
		err        error
	}{
		{"2+2*2", 6, nil},
		{"(2+2)*2", 8, nil},
		{"2+2*2+2", 8, nil},
		{"2+2*2-2", 4, nil},
		{"2+2*2/2", 4, nil},
		{"2+2*2/0", 0, errors.New("division by zero")},
		{"2+2*2a", 0, ErrInvalidExpression},
	}

	for _, tt := range tests {
		result, err := Calc(tt.expression)
		if err != tt.err && err.Error() != tt.err.Error() {
			t.Errorf("Calc(%s) error = %v, want %v", tt.expression, err, tt.err)
		}
		if result != tt.expected {
			t.Errorf("Calc(%s) = %v, want %v", tt.expression, result, tt.expected)
		}
	}
}
