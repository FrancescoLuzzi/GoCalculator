package operations

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// basic Operators

// Operator func type definition
type Operator func(x float64, y float64) (float64, error)

// Operators
func add(x float64, y float64) (float64, error) {
	return x + y, nil
}
func substract(x float64, y float64) (float64, error) {
	return x - y, nil
}
func divide(x float64, y float64) (float64, error) {
	if y == 0 {
		return 0, errors.New("can't divide by zero")
	}
	return x / y, nil
}
func multiply(x float64, y float64) (float64, error) {
	return x * y, nil
}

// function maps with corisponding sign
var FUNCTION_MAP = map[string]Operator{
	"+": add,
	"-": substract,
	"/": divide,
	"*": multiply,
}

var SUPPORTED_OPERATIONS = `
- Addition "+"
- Substraction "-"
- Division "/"
- Multiplication "*"`

// operation interface to use simple and composed operations the same way
type Operation interface {
	Execute_operation()
	Get_results() (float64, string, error)
	Set_wait_group(*sync.WaitGroup)
}

// Simple_operation type definition
type Simple_operation struct {
	Operands   []float64
	Operator   string
	wg         *sync.WaitGroup
	result     float64
	result_out string
	error      error
}

// operations result's type definition
type op_result struct {
	result     float64
	result_out string
	error      error
}

// simple Operator, given an array of floats and the operation as a string
func simple_Operator(Operands []float64, Operator_str string) op_result {
	out := new(op_result)
	out.result = Operands[0]
	out.result_out = fmt.Sprintf("(%.2f", Operands[0])
	var err error
	var op Operator
	var is_present bool
	for _, x := range Operands[1:] {
		op, is_present = FUNCTION_MAP[Operator_str]
		// if the element is not in the map
		if !is_present {
			out.error = fmt.Errorf("operand \"%s\" not supported\nSupported operations:%s", Operator_str, SUPPORTED_OPERATIONS)
			return *out
		}
		out.result, err = op(out.result, x)
		// this means division by zero
		if err != nil {
			tmp := fmt.Sprintf("%s -> (", err.Error())
			for _, y := range Operands[:len(Operands)-1] {
				tmp += fmt.Sprintf("%.2f %s", y, Operator_str)
			}
			tmp += fmt.Sprintf(" %.2f)=0", Operands[len(Operands)-1])
			out.error = errors.New(tmp)
			return *out
		}
		out.result_out += fmt.Sprintf(" %s %.2f", Operator_str, x)
	}
	out.result_out += ")"
	return *out
}

// Simple_operation's execute func definition
// it has support to be callable in a go routine,
// if provided in the Simple_operation
func (s *Simple_operation) Execute_operation() {
	result := simple_Operator(s.Operands, s.Operator)
	if result.error == nil {
		s.result_out = result.result_out
		s.result = result.result
	} else {
		s.error = result.error
		s.result = 0
	}
	if s.wg != nil {
		s.wg.Done()
	}
}

// Simple_operation's Get_results func definition
func (s *Simple_operation) Get_results() (float64, string, error) {
	return s.result, s.result_out, s.error
}

// Simple_operation's Set_wait_group func definition
func (s *Simple_operation) Set_wait_group(new_wg *sync.WaitGroup) {
	s.wg = new_wg
}

// composed operation, the operation interface
type Composed_operation struct {
	Operands   []Operation
	Operator   string
	wg         *sync.WaitGroup
	result     float64
	result_out string
	error      error
}

func (c *Composed_operation) Execute_operation() {
	result := composed_Operator(c.Operands, c.Operator)
	if result.error == nil {
		c.result = result.result
		c.result_out = result.result_out
	} else {
		c.error = result.error
		c.result = 0
	}
	if c.wg != nil {
		c.wg.Done()
	}
}

func (c *Composed_operation) Get_results() (float64, string, error) {
	return c.result, c.result_out, c.error
}

func (c *Composed_operation) Set_wait_group(new_wg *sync.WaitGroup) {
	c.wg = new_wg
}

// composed Operator
func composed_Operator(Operands []Operation, Operator_str string) op_result {
	/// execute operations in goroutines
	var curr_wg sync.WaitGroup
	out := new(op_result)
	curr_wg.Add(len(Operands))
	for _, x := range Operands {
		x.Set_wait_group(&curr_wg)
		go x.Execute_operation()
	}
	curr_wg.Wait()

	// check and aggregate all results
	var err error
	var tmp_out string
	var tmp_res float64
	tmp_res, tmp_out, err = Operands[0].Get_results()
	if err != nil {
		out.error = err
		return *out
	}
	out.result = tmp_res

	// determine which parenthesis to use
	var starter, ender string
	if strings.Contains(tmp_out, "[") || strings.Contains(tmp_out, "{") {
		starter = "{"
		ender = "}"
	} else {
		starter = "["
		ender = "]"
	}
	out.result_out = fmt.Sprintf("%s%s", starter, tmp_out)
	var op Operator
	var is_present bool
	for _, x := range Operands[1:] {
		tmp_res, tmp_out, err = x.Get_results()
		if err != nil {
			out.error = err
			return *out
		}
		op, is_present = FUNCTION_MAP[Operator_str]
		// if the element is not in the map
		if !is_present {
			out.error = fmt.Errorf("operand \"%s\" not supported\nSupported operations:%s", Operator_str, SUPPORTED_OPERATIONS)
			return *out
		}
		out.result, err = op(out.result, tmp_res)
		if err != nil {
			out.error = fmt.Errorf(("%s -> %s=0"), err.Error(), tmp_out)
			return *out
		}
		out.result_out += fmt.Sprintf("%s%s", Operator_str, tmp_out)
	}
	out.result_out += ender
	return *out
}
