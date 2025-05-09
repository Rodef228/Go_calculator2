package ast

import (
	"calculator/pkg/models"
	"testing"
)

func TestExpErr(t *testing.T) {
	tests := []struct {
		expression string
		err        error
	}{
		{"2+3", nil},
		{"2+", models.ErrInvalidExpression},
		{"+2", models.ErrOperatorFirst},
		{"2++3", models.ErrMergedOperators},
		{"2+()", models.ErrEmptyBrackets},
		{"2+)3", models.ErrNotOpenedBracket},
		{"2+3a", models.ErrWrongCharacter},
		{"2/0", models.ErrDivisionByZero},
		{"(", models.ErrInvalidExpression},
		{"", models.ErrNoOperators},
	}

	for _, tt := range tests {
		err := expErr(tt.expression)
		if err != tt.err {
			t.Errorf("expErr(%s) = %v, expected %v", tt.expression, err, tt.err)
		}
	}
}
