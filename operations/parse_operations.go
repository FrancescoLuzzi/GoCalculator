package operations

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func Parse_string(op_string string) (Operation, error) {
	chars_search := regexp.MustCompile("[a-zA-Z]")
	if chars_search.Find([]byte(op_string)) != nil {
		return nil, errors.New("character found!")
	}

	if strings.Count(op_string, "(") != strings.Count(op_string, ")") {
		return nil, errors.New("missmatching round parentheses!")
	}
	if strings.Count(op_string, "[") != strings.Count(op_string, "]") {
		return nil, errors.New("missmatching square parentheses!")
	}
	if strings.Count(op_string, "{") != strings.Count(op_string, "}") {
		return nil, errors.New("missmatching curly parentheses!")
	}
	op_string = strings.ReplaceAll(strings.Trim(op_string, " \t\n"), ",", ".")

	float_number := "(\\d+[\\.\\d+]*)"
	broken_float_number := "(\\d\\.[^\\d]+)"
	open_par := "(\\(|\\[|\\{)"
	close_par := "(\\)|\\]|\\})"
	operand := "(\\+|\\-|\\*|/)"

	reg := fmt.Sprintf("((%s%s)|(%s%s)|(%s%s$)|(^%s%s)|%s|%s$)", operand, close_par, open_par, operand, float_number, operand, operand, float_number, broken_float_number, "\\.")
	wrong_op := regexp.MustCompile(reg)

	if wrong_found := wrong_op.Find([]byte(op_string)); wrong_found != nil {
		return nil, fmt.Errorf("wrong operation format %s", wrong_found)
	}

	// just to return something
	float_seed := float64(69)
	operand1 := new(Simple_operation)
	operand1.Operands = []float64{float_seed, float_seed * 2}
	operand1.Operator = "*"
	return Operation(operand1), nil
}
