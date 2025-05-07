package calculator

import (
	"calculator/internal/service"
	"calculator/pkg/config"
	"calculator/pkg/task"
	"errors"
	"strconv"
	"unicode"
)

type tokenType int

const (
	tokenNumber tokenType = iota
	tokenOperator
	tokenLParen
	tokenRParen
)

type token struct {
	typ tokenType
	val string
}

type Calculator struct {
	taskService *service.TaskService
}

func New(taskService *service.TaskService) *Calculator {
	return &Calculator{
		taskService: taskService,
	}
}

func (c *Calculator) Calculate(expr string) (float64, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}

	rpn, err := shuntingYard(tokens)
	if err != nil {
		return 0, err
	}

	return c.evalRPN(rpn)
}

func tokenize(expr string) ([]token, error) {
	var tokens []token
	i := 0
	for i < len(expr) {
		ch := expr[i]
		if ch == ' ' {
			i++
			continue
		}

		if unicode.IsDigit(rune(ch)) || ch == '.' {
			start := i
			dotCount := 0
			for i < len(expr) && (unicode.IsDigit(rune(expr[i])) || expr[i] == '.') {
				if expr[i] == '.' {
					dotCount++
					if dotCount > 1 {
						return nil, errors.New("invalid number format")
					}
				}
				i++
			}
			tokens = append(tokens, token{typ: tokenNumber, val: expr[start:i]})
		} else if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			tokens = append(tokens, token{typ: tokenOperator, val: string(ch)})
			i++
		} else if ch == '(' {
			tokens = append(tokens, token{typ: tokenLParen, val: string(ch)})
			i++
		} else if ch == ')' {
			tokens = append(tokens, token{typ: tokenRParen, val: string(ch)})
			i++
		} else {
			return nil, errors.New("invalid character: " + string(ch))
		}
	}
	return tokens, nil
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

func shuntingYard(tokens []token) ([]token, error) {
	var output []token
	var stack []token

	for _, tok := range tokens {
		switch tok.typ {
		case tokenNumber:
			output = append(output, tok)
		case tokenOperator:
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.typ == tokenOperator && precedence(top.val) >= precedence(tok.val) {
					output = append(output, top)
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, tok)
		case tokenLParen:
			stack = append(stack, tok)
		case tokenRParen:
			foundLParen := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.typ == tokenLParen {
					foundLParen = true
					break
				} else {
					output = append(output, top)
				}
			}
			if !foundLParen {
				return nil, errors.New("bracket mismatch")
			}
		}
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.typ == tokenLParen || top.typ == tokenRParen {
			return nil, errors.New("bracket mismatch")
		}
		output = append(output, top)
	}

	return output, nil
}

func (c *Calculator) evalRPN(tokens []token) (float64, error) {
	var stack []*task.Future

	for _, tok := range tokens {
		switch tok.typ {
		case tokenNumber:
			num, err := strconv.ParseFloat(tok.val, 64)
			if err != nil {
				return 0, err
			}
			future := task.NewFuture()
			future.SetResult(num)
			stack = append(stack, future)
		case tokenOperator:
			if len(stack) < 2 {
				return 0, errors.New("invalid expression")
			}

			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			future := task.NewFuture()
			stack = append(stack, future)

			task := task.Task{
				Arg1:      a.Get(),
				Arg2:      b.Get(),
				Operation: tok.val,
			}

			// Create task and wait for result
			createdTask := c.taskService.CreateTask(task.Arg1, task.Arg2, task.Operation, getOperationTime(tok.val))
			c.taskService.SetResult(createdTask.ID, future.Get())
		}
	}

	if len(stack) != 1 {
		return 0, errors.New("invalid expression")
	}

	return stack[0].Get(), nil
}

func getOperationTime(op string) int {
	switch op {
	case "+":
		return config.Configuration.AddTimeMs
	case "-":
		return config.Configuration.SubtractTimeMs
	case "*":
		return config.Configuration.MultiplyTimeMs
	case "/":
		return config.Configuration.DivideTimeMs
	default:
		return 0
	}
}
