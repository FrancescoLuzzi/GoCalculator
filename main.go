package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
)

const NO_BASE_CMD_ERROR = 1
const WRONG_MULTI_CMD_ERROR = 2

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

var FUNCTION_MAP = map[byte]func(float64, float64) (float64, error){'+': add, '-': substract, '/': divide, '*': multiply}

type op_result struct {
	result     float64
	result_out string
	error      error
}

type base_operation[IN any] func(operands []IN) op_result

// simple operation
type simple_operation struct {
	operands   []float64
	operator   base_operation[float64]
	wg         *sync.WaitGroup
	result     float64
	result_out string
	error      error
}

func (s *simple_operation) execute_operation() {
	result := s.operator(s.operands)
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

// simple operator
func simple_add(operands []float64) op_result {
	out := new(op_result)
	out.result = operands[0]
	out.result_out = fmt.Sprintf("(%.2f", operands[0])
	for _, x := range operands[1:] {
		out.result += x
		out.result_out += fmt.Sprintf("+%.2f", x)
	}
	out.result_out += ")"
	return *out
}
func simple_substract(operands []float64) op_result {
	out := new(op_result)
	out.result = operands[0]
	out.result_out = fmt.Sprintf("(%.2f", operands[0])
	for _, x := range operands[1:] {
		out.result -= x
		out.result_out += fmt.Sprintf("-%.2f", x)
	}
	out.result_out += ")"
	return *out
}
func simple_multiply(operands []float64) op_result {
	out := new(op_result)
	out.result = operands[0]
	out.result_out = fmt.Sprintf("(%.2f", operands[0])
	for _, x := range operands[1:] {
		out.result *= x
		out.result_out += fmt.Sprintf("*%.2f", x)
	}
	out.result_out += ")"
	return *out
}
func simple_divide(operands []float64) op_result {
	out := new(op_result)
	out.result = operands[0]
	out.result_out = fmt.Sprintf("(%.2f", operands[0])
	for _, x := range operands[1:] {
		if x == 0 {
			out.result = 0
			tmp := "Error, can't divide by 0!\n{ "
			for _, y := range operands[:len(operands)-1] {
				tmp += fmt.Sprintf("%.2f, ", y)
			}
			tmp += fmt.Sprintf("%.2f }", operands[len(operands)-1])
			out.error = errors.New(tmp)
			return *out
		}
		out.result += x
		out.result_out += fmt.Sprintf("/%.2f", x)
	}
	out.result_out += ")"
	return *out
}

// composed operation
type composed_operation struct {
	operands   []*simple_operation
	operator   base_operation[*simple_operation]
	wg         *sync.WaitGroup
	result     float64
	result_out string
	error      error
}

func (s *composed_operation) execute_operation() {
	result := s.operator(s.operands)
	if result.error == nil {
		s.result = result.result
		s.result_out = result.result_out
	} else {
		s.error = result.error
		s.result = 0
	}
	if s.wg != nil {
		s.wg.Done()
	}
}

// composed operator
func composed_add(operands []simple_operation) op_result {
	out := new(op_result)
	operands[0].execute_operation()
	if operands[0].error != nil {
		out.error = operands[0].error
		return *out
	}
	out.result = operands[0].result
	out.result_out = fmt.Sprintf("[%s", operands[0].result_out)
	for _, x := range operands[1:] {
		x.execute_operation()
		if x.error != nil {
			out.error = x.error
			return *out
		}
		out.result += x.result
		out.result_out = fmt.Sprintf("+%s", x.result_out)
	}
	out.result_out += "]"
	return *out
}
func composed_substract(operands []simple_operation) op_result {
	out := new(op_result)
	operands[0].execute_operation()
	if operands[0].error != nil {
		out.error = operands[0].error
		return *out
	}
	out.result = operands[0].result
	out.result_out = fmt.Sprintf("[%s", operands[0].result_out)
	for _, x := range operands[1:] {
		x.execute_operation()
		if x.error != nil {
			out.error = x.error
			return *out
		}
		out.result -= x.result
		out.result_out = fmt.Sprintf("-%s", x.result_out)
	}
	out.result_out += "]"
	return *out
}
func composed_multiply(operands []simple_operation) op_result {
	out := new(op_result)
	operands[0].execute_operation()
	if operands[0].error != nil {
		out.error = operands[0].error
		return *out
	}
	out.result = operands[0].result
	out.result_out = fmt.Sprintf("[%s", operands[0].result_out)
	for _, x := range operands[1:] {
		x.execute_operation()
		if x.error != nil {
			out.error = x.error
			return *out
		}
		out.result *= x.result
		out.result_out = fmt.Sprintf("*%s", x.result_out)
	}
	out.result_out += "]"
	return *out
}
func composed_divide(operands []simple_operation) op_result {
	out := new(op_result)
	operands[0].execute_operation()
	if operands[0].error != nil {
		out.error = operands[0].error
		return *out
	}
	out.result = operands[0].result
	out.result_out = fmt.Sprintf("[%s", operands[0].result_out)
	for _, x := range operands[1:] {
		x.execute_operation()
		if x.error != nil {
			out.error = x.error
			return *out
		} else if x.result == 0 {
			out.result = 0
			tmp := "Error, can divide by 0!\n{ "
			for _, y := range x.operands[:len(x.operands)-1] {
				tmp += fmt.Sprintf("%.2f, ", y)
			}
			tmp += fmt.Sprintf("%.2f }", x.operands[len(x.operands)-1])
			out.error = errors.New(tmp)
			return *out
		}
		out.result /= x.result
		out.result_out = fmt.Sprintf("/%s", x.result_out)
	}
	out.result_out += "]"
	return *out
}

func handle_multiple_workers(multiCmd *flag.FlagSet, number_of_workers *int) {
	// if not enough arguments
	if len(os.Args) < 3 {
		multiCmd.PrintDefaults()
		os.Exit(WRONG_MULTI_CMD_ERROR)
	}

	//parse cmds
	multiCmd.Parse(os.Args[2:])

	// if number of operations/workers is zero exit
	if *number_of_workers == 0 {
		multiCmd.PrintDefaults()
		os.Exit(WRONG_MULTI_CMD_ERROR)
	}

	operators := [4]base_operation[float64]{simple_add, simple_substract, simple_multiply, simple_divide}
	tot_operators := len(operators)
	operations := make([]simple_operation, *number_of_workers)

	var goroutines_wg sync.WaitGroup
	goroutines_wg.Add(*number_of_workers)
	var curr_op base_operation[float64]
	for i := 1; i <= *number_of_workers; i++ {
		curr_op = operators[i%tot_operators]
		operations[i-1] = simple_operation{[]float64{float64(i * i), float64(i) * rand.Float64()}, curr_op, &goroutines_wg, 0, "", nil}
		go operations[i-1].execute_operation()
	}
	goroutines_wg.Wait()
	for i, op := range operations {
		if op.error == nil {
			fmt.Printf("Go-routine %d -> %s\n", i+1, op.result_out)
		} else {
			fmt.Printf("Go-routine %d Error -> %s\n", i+1, op.error)
		}
	}
}

func main() {
	// multile workers usage
	multiCmd := flag.NewFlagSet("multi", flag.ExitOnError)
	number_of_workers := multiCmd.Int("number", 0, "Number of operations done, each operation is done in a goroutine.\nThis must be >0")

	if len(os.Args) < 2 {
		fmt.Println("You need to enter a basic comand:\n-multi\n-simple\n-from_file")
		os.Exit(NO_BASE_CMD_ERROR)
	}

	switch os.Args[1] {
	case "multi":
		handle_multiple_workers(multiCmd, number_of_workers)
	case "simple":
		fmt.Println("simple CMD")

	case "from_file":
		fmt.Println("from_file CMD")
	default:
		fmt.Println("You need to enter a basic comand:\n-multi\n-simple\n-from_file")
		os.Exit(NO_BASE_CMD_ERROR)
	}

}
