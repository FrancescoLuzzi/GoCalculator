package operations

import (
	"fmt"
	"testing"
)

var operators = [4]string{"+", "-", "/", "*"}
var numbers = [6]string{"3.1415", "3", "3.", "7,9", "g", "O"}

func TestWrongParseSimple(t *testing.T) {
	base_op := "%s%s%s%s"
	current_op_test := 0
	for _, n := range numbers {
		for _, o_par := range open_par {
			for _, c_par := range close_par {
				for _, oper := range operators {
					current_op_test++
					op := fmt.Sprintf(base_op, o_par, oper, n, c_par)
					_, err := Parse_string(op)
					if err == nil {
						t.Errorf("%d -> %s did not errored out", current_op_test, op)
					}
					t.Logf("%d -> tested operation %s -> %s\n", current_op_test, op, err)
					current_op_test++
					op = fmt.Sprintf(base_op, o_par, n, oper, c_par)
					_, err = Parse_string(op)
					if err == nil {
						t.Errorf("%d -> %s did not errored out", current_op_test, op)
					}
					t.Logf("%d -> tested operation %s -> %s\n", current_op_test, op, err)
				}
			}
		}
	}
}

func TestWrongParseCustom(t *testing.T) {
	op := "4*3."
	_, err := Parse_string(op)
	if err == nil {
		t.Errorf("%s did not errored out", op)
	} else {
		t.Logf("%s did errored out -> %s\n", op, err)
	}
}
