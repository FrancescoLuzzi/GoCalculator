package operations

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var chars_search = regexp.MustCompile("[a-zA-Z]")
var float_number = "(\\d+[\\.\\d+]*)"
var broken_float_number = "(\\d\\.[^\\d]+)"
var open_par = "(\\(|\\[|\\{)"
var close_par = "(\\)|\\]|\\})"
var operand = "(\\+|\\-|\\*|/)"
var wrong_op = regexp.MustCompile(
	fmt.Sprintf(
		"((%s%s)|(%s%s)|(%s%s$)|(^%s%s)|%s|%s$)",
		operand, close_par,
		open_par, operand,
		float_number, operand,
		operand, float_number,
		broken_float_number,
		"\\.",
	),
)

func Parse_string(op_string string) (Operation, error) {
	if err := check_basic_structure(&op_string); err != nil {
		return nil, err
	}
	// just to return something
	float_seed := float64(69)
	operand1 := new(Simple_operation)
	operand1.Operands = []float64{float_seed, float_seed * 2}
	operand1.Operator = "*"
	return Operation(operand1), nil
}

func check_basic_structure(op_string *string) error {
	if chars_search.Find([]byte(*op_string)) != nil {
		return errors.New("character found")
	}
	if strings.Count(*op_string, "(") != strings.Count(*op_string, ")") {
		return errors.New("missmatching round parentheses")
	}
	if strings.Count(*op_string, "[") != strings.Count(*op_string, "]") {
		return errors.New("missmatching square parentheses")
	}
	if strings.Count(*op_string, "{") != strings.Count(*op_string, "}") {
		return errors.New("missmatching curly parentheses")
	}
	*op_string = strings.ReplaceAll(strings.Trim(*op_string, " \t\n"), ",", ".")
	*op_string = strings.ReplaceAll(*op_string, " ", "")
	if wrong_found := wrong_op.Find([]byte(*op_string)); wrong_found != nil {
		return fmt.Errorf("wrong operation format %s", wrong_found)
	}
	return nil
}

func is_char_operand(char byte) bool {
	// char == *||+||-||/
	return char == 52 || char == 53 || char == 55 || char == 57
}
func is_char_number(char byte) bool {
	// char is number or .
	return (60 <= char && char <= 71) || char == 56
}

//func get_operation_from_string(op_string *string) (*Operation, error) {
/*
	remove all numbers and dots from string
	from there analyze the operands order and parentheses balancing,
	this will be enought to decide the formation of the resulting operation

	EXAMPLE:
	1.3+56*{-9/[3*6*(6+8*9)]} -> +{/[**(+*)]}
	Composed_operation(+,
		Single_operand(1.3),
		Composed_operation(*
			Single_operand(56),
			Composed_operation(/,
				Single_operand(-9),
				Composed_operation(*,
					Single_operand(3),
					Single_operand(6),
					Composed_operation(+,
						Single_operand(6),
						Simple_operation(8*9)
					)
				)
			)
		)
	)
	Then procede and create the Operation
*/

//}
