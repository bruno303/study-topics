package lang

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func Evaluate(input string, env *Environment) (any, error) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return nil, nil
	}

	switch tokens[0] {
	case "let":
		if tokens[2] != "=" {
			return nil, fmt.Errorf("syntax error. Usage: let x = 10")
		}

		var val float64
		var err error
		var ok bool

		if len(tokens) == 4 {
			val, err = strconv.ParseFloat(tokens[3], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number")
			}
		} else if len(tokens) > 4 {
			expResult, err := evalExpression(tokens[3:], env)
			if err != nil {
				return nil, err
			}
			val, ok = expResult.(float64)
			if !ok {
				return nil, fmt.Errorf("invalid number")
			}
		} else {
			return nil, fmt.Errorf("syntax error. Usage: let x = 10")
		}

		env.Set(tokens[1], val)
		return nil, nil

	case "print":
		if len(tokens) == 2 {
			// Print variable
			val, ok := env.Get(tokens[1])
			if !ok {
				return nil, fmt.Errorf("undefined variable: %s", tokens[1])
			}
			return val, nil
		} else {
			// Print expression
			result, err := evalExpression(tokens[1:], env)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	default:
		// Try evaluating an arithmetic expression like: x + 5
		return evalExpression(tokens, env)
	}
}

func evalExpression(tokens []string, env *Environment) (any, error) {
	if len(tokens) != 3 {
		return 0, fmt.Errorf("expected expression like: x + 5 or x == 5")
	}

	leftVal, err := resolveOperand(tokens[0], env)
	if err != nil {
		return 0, err
	}

	rightVal, err := resolveOperand(tokens[2], env)
	if err != nil {
		return 0, err
	}

	switch tokens[1] {
	case "+":
		return leftVal + rightVal, nil
	case "-":
		return leftVal - rightVal, nil
	case "*":
		return leftVal * rightVal, nil
	case "/":
		return leftVal / rightVal, nil
	case "^":
		return math.Pow(leftVal, rightVal), nil

	case "==":
		return leftVal == rightVal, nil
	case "!=":
		return leftVal != rightVal, nil
	case ">":
		return leftVal > rightVal, nil
	case "<":
		return leftVal < rightVal, nil
	case ">=":
		return leftVal >= rightVal, nil
	case "<=":
		return leftVal <= rightVal, nil

	default:
		return 0, fmt.Errorf("unsupported operator: %s", tokens[1])
	}
}

func resolveOperand(token string, env *Environment) (float64, error) {
	val, err := strconv.ParseFloat(token, 64)
	if err == nil {
		return val, nil
	}

	if v, ok := env.Get(token); ok {
		return v, nil
	}

	return 0, fmt.Errorf("unknown variable or value: %s", token)
}
