package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"

	"github.com/FrancescoLuzzi/GoCalculator/operations"
)

const NO_BASE_CMD_ERROR = 1
const WRONG_MULTI_CMD_ERROR = 2

// loggers
var (
	INFO_LOGGER    *log.Logger
	WARNING_LOGGER *log.Logger
	ERROR_LOGGER   *log.Logger
)

// init loggers
func init_loggers(on_file bool) {
	INFO_LOGGER = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	WARNING_LOGGER = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	ERROR_LOGGER = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	if on_file {
		file, err := os.OpenFile("calculator_log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			ERROR_LOGGER.Fatal("Cannot open calculator_log.txt file")
		}
		INFO_LOGGER.SetOutput(file)
		WARNING_LOGGER.SetOutput(file)
		ERROR_LOGGER.SetOutput(file)
	}
}

// given array of operations print the output
func print_output(operations []operations.Operation) {
	var res float64
	var out string
	var err error
	for i, op := range operations {
		res, out, err = op.Get_results()
		if err == nil {
			INFO_LOGGER.Printf("Go-routine %d -> %s=%.2f\n", i+1, out, res)
		} else {
			ERROR_LOGGER.Printf("Go-routine %d Error -> %s\n", i+1, err)
		}
	}
}

// generate a operations.Composed_operation rapidly
func generate_composed_operation(seed int, wg *sync.WaitGroup) *operations.Composed_operation {
	float_seed := float64(seed)
	operand1 := new(operations.Simple_operation)
	operand1.Operands = []float64{float_seed, float_seed * 2}
	operand1.Operator = "*"

	operand2 := new(operations.Simple_operation)
	operand2.Operands = []float64{float_seed * 2, float_seed * 3}
	operand2.Operator = "+"

	comp_op1 := new(operations.Composed_operation)
	comp_op1.Operands = []operations.Operation{operand1, operand2}
	comp_op1.Operator = "/"

	operand3 := new(operations.Simple_operation)
	operand3.Operands = []float64{float_seed * 3, 3}
	operand3.Operator = "+"

	operand4 := new(operations.Simple_operation)
	operand4.Operands = []float64{float_seed, 2}
	operand4.Operator = "/"

	comp_op2 := new(operations.Composed_operation)
	comp_op2.Operands = []operations.Operation{operand3, operand4}
	comp_op2.Operator = "-"

	single_op := new(operations.Single_operand)
	single_op.Value = 7.0

	comp_op := new(operations.Composed_operation)
	comp_op.Operands = []operations.Operation{comp_op1, comp_op2, single_op}
	comp_op.Operator = "+"
	comp_op.Set_wait_group(wg)
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
	operations_to_do := make([]operations.Operation, *number_of_operations)
	var goroutines_wg sync.WaitGroup
	goroutines_wg.Add(*number_of_operations)
	if *is_composed {
		for i := 1; i <= *number_of_operations; i++ {
			operations_to_do[i-1] = operations.Operation(generate_composed_operation(i, &goroutines_wg))
			go operations_to_do[i-1].Execute_operation()
		}
	} else {
		var tmp_op *operations.Simple_operation
		for i := 1; i <= *number_of_operations; i++ {
			tmp_op = new(operations.Simple_operation)
			tmp_op.Operands = []float64{float64(i * i), float64(i) * rand.Float64()}
			tmp_op.Operator = operators[i%tot_operators]
			operations_to_do[i-1] = operations.Operation(tmp_op)
			operations_to_do[i-1].Set_wait_group(&goroutines_wg)
			go operations_to_do[i-1].Execute_operation()
		}
	}
	goroutines_wg.Wait()
	print_output(operations_to_do)

}

// init app
func init() {
	init_loggers(false)
}

func main() {
	// multile workers usage
	multiCmd := flag.NewFlagSet("multi", flag.ExitOnError)
	number_of_operations := multiCmd.Int("number", 0, "Number of operations done, the operations are done in a goroutine.\nThis must be >0")
	is_composed := multiCmd.Bool("composed", false, "Determine if operations are composed")
	if len(os.Args) < 1 {
		fmt.Println("You need to enter a basic comand:\n- multi\n- simple\n- from_file")
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
