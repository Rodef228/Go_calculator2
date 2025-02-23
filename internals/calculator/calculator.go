package calculator

import (
	"errors"
	"http_calculator/internals/config"
	"strings"
	"time"
)

var ErrInvalidExpression = errors.New("invalid expression format")
var cfg = config.Load()

func priority(operator rune) int {
	switch operator {
	case '*', '/':
		return 2
	case '+', '-':
		return 1
	default:
		return 0
	}
}

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	operators := []rune{}
	nums := []float64{}

	for i := 0; i < len(expression); i++ {
		chr := expression[i]

		if chr >= '0' && chr <= '9' {
			num := float64(0)
			for i < len(expression) && expression[i] >= '0' && expression[i] <= '9' {
				num = num*10 + float64(expression[i]-'0')
				i++
			}
			nums = append(nums, num)
			i--
		} else if chr == '(' {
			operators = append(operators, '(')
		} else if chr == ')' {
			for len(operators) > 0 && operators[len(operators)-1] != '(' {
				if err := MiniCalc(&nums, &operators); err != nil {
					return 0, err
				}
			}
			if len(operators) > 0 {
				operators = operators[:len(operators)-1]
			} else {
				return 0, ErrInvalidExpression
			}
		} else if chr == '+' || chr == '-' || chr == '*' || chr == '/' {
			for len(operators) > 0 && priority(operators[len(operators)-1]) >= priority(rune(chr)) {
				if err := MiniCalc(&nums, &operators); err != nil {
					return 0, err
				}
			}
			operators = append(operators, rune(chr))
		} else {
			return 0, ErrInvalidExpression
		}
	}

	for len(operators) > 0 {
		if err := MiniCalc(&nums, &operators); err != nil {
			return 0, err
		}
	}

	if len(nums) != 1 {
		return 0, ErrInvalidExpression
	}

	return nums[0], nil
}

func MiniCalc(nums *[]float64, operators *[]rune) error {
	if len(*nums) < 2 || len(*operators) == 0 {
		return ErrInvalidExpression
	}

	operator := (*operators)[len(*operators)-1]
	*operators = (*operators)[:len(*operators)-1]
	right := (*nums)[len(*nums)-1]
	left := (*nums)[len(*nums)-2]
	*nums = (*nums)[:len(*nums)-2]

	switch operator {
	case '+':
		*nums = append(*nums, left+right)
		time.Sleep(cfg.TIME_ADDITION_MS)
	case '-':
		*nums = append(*nums, left-right)
		time.Sleep(cfg.TIME_SUBTRACTION_MS)
	case '*':
		*nums = append(*nums, left*right)
		time.Sleep(cfg.TIME_MULTIPLICATIONS_MS)
	case '/':
		if right == 0 {
			return errors.New("division by zero")
		}
		*nums = append(*nums, left/right)
		time.Sleep(cfg.TIME_DIVISIONS_MS)
	}
	return nil
}
