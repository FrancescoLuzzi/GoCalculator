package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
)

const NO_BASE_CMD_ERROR = 1
const WRONG_MULTI_CMD_ERROR = 2

// basic operators

// operator func type definition
type operator func(x float64, y float64) (float64, error)

// operators
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
var FUNCTION_MAP = map[string]operator{
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
type operation interface {
	execute_operation()
	get_results() (float64, string, error)
	set_wait_group(*sync.WaitGroup)
}

// simple_operation type definition
type simple_operation struct {
	operands   []float64
	operator   string
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

// simple operator, given an array of floats and the operation as a string
func simple_operator(operands []float64, operator_str string) op_result {
	out := new(op_result)
	out.result = operands[0]
	out.result_out = fmt.Sprintf("(%.2f", operands[0])
	var err error
	var op operator
	var is_present bool
	for _, x := range operands[1:] {
		op, is_present = FUNCTION_MAP[operator_str]
		// if the element is not in the map
		if !is_present {
			out.error = fmt.Errorf("operand \"%s\" not supported\nSupported operations:%s", operator_str, SUPPORTED_OPERATIONS)
			return *out
		}
		out.result, err = op(out.result, x)
		// this means division by zero
		if err != nil {
			tmp := fmt.Sprintf("%s\n(", err.Error())
			for _, y := range operands[:len(operands)-1] {
				tmp += fmt.Sprintf("%.2f %s", y, operator_str)
			}
			tmp += fmt.Sprintf(" %.2f)=0", operands[len(operands)-1])
			out.error = errors.New(tmp)
			return *out
		}
		out.result_out += fmt.Sprintf(" %s %.2f", operator_str, x)
	}
	out.result_out += ")"
	return *out
}

// simple_operation's execute func definition
// it has support to be callable in a go routine,
// if provided in the simple_operation
func (s *simple_operation) execute_operation() {
	result := simple_operator(s.operands, s.operator)
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

// simple_operation's get_results func definition
func (s *simple_operation) get_results() (float64, string, error) {
	return s.result, s.result_out, s.error
}

// simple_operation's set_wait_group func definition
func (s *simple_operation) set_wait_group(new_wg *sync.WaitGroup) {
	s.wg = new_wg
}

// composed operation, the operation interface
type composed_operation struct {
	operands   []operation
	operator   string
	wg         *sync.WaitGroup
	result     float64
	result_out string
	error      error
}

func (c *composed_operation) execute_operation() {
	result := composed_operand(c.operands, c.operator)
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

func (c *composed_operation) get_results() (float64, string, error) {
	return c.result, c.result_out, c.error
}

func (c *composed_operation) set_wait_group(new_wg *sync.WaitGroup) {
	c.wg = new_wg
}

// composed operator
func composed_operand(operands []operation, operator_str string) op_result {
	/// execute operations in goroutines
	var curr_wg sync.WaitGroup
	out := new(op_result)
	curr_wg.Add(len(operands))
	for _, x := range operands {
		x.set_wait_group(&curr_wg)
		go x.execute_operation()
	}
	curr_wg.Wait()

	// check and aggregate all results
	var err error
	var tmp_out string
	var tmp_res float64
	tmp_res, tmp_out, err = operands[0].get_results()
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
	var op operator
	var is_present bool
	for _, x := range operands[1:] {
		tmp_res, tmp_out, err = x.get_results()
		if err != nil {
			out.error = err
			return *out
		}
		op, is_present = FUNCTION_MAP[operator_str]
		// if the element is not in the map
		if !is_present {
			out.error = fmt.Errorf("operand \"%s\" not supported\nSupported operations:%s", operator_str, SUPPORTED_OPERATIONS)
			return *out
		}
		out.result, err = op(out.result, tmp_res)
		if err != nil {
			tmp := fmt.Sprintln(err.Error())
			tmp += tmp_out
			tmp += "=0"
			out.error = errors.New(tmp)
			return *out
		}
		out.result_out += fmt.Sprintf("%s%s", operator_str, tmp_out)
	}
	out.result_out += ender
	return *out
}

// given array of operations print the output
func print_output(operations []operation) {
	var res float64
	var out string
	var err error
	for i, op := range operations {
		res, out, err = op.get_results()
		if err == nil {
			fmt.Printf("Go-routine %d -> %s=%.2f\n", i+1, out, res)
		} else {
			fmt.Printf("Go-routine %d Error -> %s\n", i+1, err)
		}
	}
}

// generate a composed_operation rapidly
func generate_composed_operation(seed int, wg *sync.WaitGroup) *composed_operation {
	float_seed := float64(seed)
	operand1 := new(simple_operation)
	operand1.operands = []float64{float_seed, float_seed * 2}
	operand1.operator = "*"

	operand2 := new(simple_operation)
	operand2.operands = []float64{float_seed * 2, float_seed * 3}
	operand2.operator = "+"

	comp_op1 := new(composed_operation)
	comp_op1.operands = []operation{operand1, operand2}
	comp_op1.operator = "/"

	operand3 := new(simple_operation)
	operand3.operands = []float64{float_seed * 3, 3}
	operand3.operator = "+"

	operand4 := new(simple_operation)
	operand4.operands = []float64{float_seed, 2}
	operand4.operator = "/"

	comp_op2 := new(composed_operation)
	comp_op2.operands = []operation{operand3, operand4}
	comp_op2.operator = "-"

	comp_op := new(composed_operation)
	comp_op.operands = []operation{comp_op1, comp_op2}
	comp_op.operator = "+"
	comp_op.wg = wg
	return comp_op
}

func handle_multiple_workers(multiCmd *flag.FlagSet, number_of_operations *int, is_composed *bool) {
	// if not enough arguments
	if len(os.Args) < 3 {
		multiCmd.PrintDefaults()
		os.Exit(WRONG_MULTI_CMD_ERROR)
	}

	//parse cmds
	multiCmd.Parse(os.Args[2:])

	// if number of operations/workers is zero exit
	if *number_of_operations == 0 {
		multiCmd.PrintDefaults()
		os.Exit(WRONG_MULTI_CMD_ERROR)
	}

	operators := [4]string{"+", "-", "*", "/"}
	tot_operators := len(operators)
	operations := make([]operation, *number_of_operations)
	var goroutines_wg sync.WaitGroup
	goroutines_wg.Add(*number_of_operations)
	if *is_composed {
		for i := 1; i <= *number_of_operations; i++ {
			operations[i-1] = operation(generate_composed_operation(i, &goroutines_wg))
			go operations[i-1].execute_operation()
		}
	} else {
		for i := 1; i <= *number_of_operations; i++ {
			operations[i-1] = operation(&simple_operation{[]float64{float64(i * i), float64(i) * rand.Float64()}, operators[i%tot_operators], &goroutines_wg, 0, "", nil})
			go operations[i-1].execute_operation()
		}
	}
	goroutines_wg.Wait()
	print_output(operations)

}

func main() {
	// multile workers usage
	multiCmd := flag.NewFlagSet("multi", flag.ExitOnError)
	number_of_operations := multiCmd.Int("number", 0, "Number of operations done, the operations are done in a goroutine.\nThis must be >0")
	is_composed := multiCmd.Bool("composed", false, "Determine if operations are composed")
	if len(os.Args) < 2 {
		fmt.Println("You need to enter a basic comand:\n-multi\n-simple\n-from_file")
		os.Exit(NO_BASE_CMD_ERROR)
	}

	switch os.Args[1] {
	case "multi":
		handle_multiple_workers(multiCmd, number_of_operations, is_composed)
		return
	case "simple":
		fmt.Println("simple CMD")
		return
	case "from_file":
		fmt.Println("from_file CMD")
		return
	default:
		fmt.Println("You need to enter a basic comand:\n-multi\n-simple\n-from_file")
		os.Exit(NO_BASE_CMD_ERROR)
	}

}
