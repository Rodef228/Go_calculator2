package ast

import "calculator/pkg/models"

// первоначальная проверка на ошибки
// понижает шанс пропустить ошибку в выражении
func expErr(expression string) error {
	len := len(expression)
	flag := false
	start := 0
	end := 0

	for i := 0; i < len; i++ {
		curr := expression[i]
		next := byte(0)
		if i < len-1 {
			next = expression[i+1]
		}

		if curr == '(' {
			start++
		}
		if curr == ')' {
			end++
		}
		if 48 <= curr && curr <= 57 && !flag {
			flag = true
		}

		switch {
		case i == 0 && (curr == ')' || curr == '*' || curr == '+' || curr == '-' || curr == '/'):
			return models.ErrOperatorFirst
		case i == len-1 && (curr == '*' || curr == '+' || curr == '-' || curr == '/'):
			return models.ErrOperatorLast
		case curr == '(' && next == ')':
			return models.ErrEmptyBrackets
		case curr == ')' && next == '(':
			return models.ErrMergedBrackets
		case (curr == '*' || curr == '+' || curr == '-' || curr == '/') && (next == '*' || next == '+' || next == '-' || next == '/'):
			return models.ErrMergedOperators
		case curr < '(' || curr > '9':
			return models.ErrWrongCharacter
		case len <= 2:
			return models.ErrInvalidExpression
		case curr == '/' && next == '0':
			return models.ErrDivisionByZero
		}
	}

	// базовая проверка на корректность скобок
	if start > end {
		return models.ErrNotClosedBracket
	} else if end > start {
		return models.ErrNotOpenedBracket
	}

	if !flag {
		return models.ErrNoOperators
	}
	return nil
}
