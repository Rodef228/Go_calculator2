package ast

import "calculator/pkg/models"

type stack []*token

func (s *stack) push(t *token) {
	*s = append(*s, t)
}

func (s *stack) pop() (*token, error) {
	if len(*s) == 0 {
		return nil, models.ErrEmptyStack
	}
	t := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return t, nil
}

func (s *stack) peek() *token {
	if len(*s) == 0 {
		return nil
	}
	return (*s)[len(*s)-1]
}

func (s *stack) len() int {
	return len(*s)
}

func rpn(tokens []*token) ([]*token, error) {
	var stack stack
	output := make([]*token, 0)

	for _, tok := range tokens {
		switch tok.t {
		case models.Operand:
			output = append(output, tok)

		case models.Operator:
			currPriority, err := priority(tok.val)
			if err != nil {
				return nil, err
			}

			// извлекаем операторы с большим или равным приоритетом
			for stack.len() > 0 {
				top := stack.peek()
				if top.t == models.OpenBracket {
					break // открывающая скобка прерывает извлечение
				}

				topPriority, err := priority(top.val)
				if err != nil {
					return nil, err
				}

				if topPriority >= currPriority {
					popped, _ := stack.pop()
					output = append(output, popped)
				} else {
					break
				}
			}
			stack.push(tok)

		case models.OpenBracket:
			stack.push(tok)

		case models.CloseBracket:
			// извлекаем до открывающей скобки
			found := false
			for stack.len() > 0 {
				popped, err := stack.pop()
				if err != nil {
					return nil, models.ErrInvalidExpression
				}
				if popped.t == models.OpenBracket {
					found = true
					break
				}
				output = append(output, popped)
			}
			if !found {
				return nil, models.ErrNotOpenedBracket
			}

		default:
			return nil, models.ErrUnknownOperator
		}
	}

	// достаем оставшиеся операторы
	for stack.len() > 0 {
		popped, err := stack.pop()
		if err != nil {
			return nil, err
		}
		if popped.t == models.OpenBracket {
			return nil, models.ErrNotClosedBracket
		}
		output = append(output, popped)
	}

	return output, nil
}
