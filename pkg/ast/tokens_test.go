package ast

import (
	"calculator/pkg/models"
	"testing"
)

func TestTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []*token
	}{
		{
			input: "2+3",
			expected: []*token{
				{t: models.Operand, val: "2"},
				{t: models.Operator, val: "+"},
				{t: models.Operand, val: "3"},
			},
		},
		{
			input: "2*(3+4)",
			expected: []*token{
				{t: models.Operand, val: "2"},
				{t: models.Operator, val: "*"},
				{t: models.OpenBracket, val: "("},
				{t: models.Operand, val: "3"},
				{t: models.Operator, val: "+"},
				{t: models.Operand, val: "4"},
				{t: models.CloseBracket, val: ")"},
			},
		},
	}

	for _, tt := range tests {
		result := tokens(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("tokens(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
		for i := range result {
			if result[i].t != tt.expected[i].t || result[i].val != tt.expected[i].val {
				t.Errorf("tokens(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		}
	}
}
