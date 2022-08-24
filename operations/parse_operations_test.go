package operations

import (
	"fmt"
	"testing"
)

var operators = [4]string{"+", "-", "/", "*"}
var open_par = [4]string{"", "(", "[", "{"}
var close_par = [4]string{"", ")", "]", "}"}
var numbers = [5]string{"3.1415", "3", "7,9", "g", "O"}

func TestWrongParseSimple(t *testing.T) {
	base_op := "%s%s%s%s"
	for _, n := range numbers {
		for _, o_par := range open_par {
			for _, c_par := range close_par {
				for _, oper := range operators {
					op := fmt.Sprintf(base_op, o_par, oper, n, c_par)
					_, err := Parse_string(op)
					if err == nil {
						t.Errorf("%s did not errored out", op)
					}
					t.Logf("tested operation %s -> %s\n", op, err)
					op = fmt.Sprintf(base_op, o_par, n, oper, c_par)
					_, err = Parse_string(op)
					if err == nil {
						t.Errorf("%s did not errored out", op)
					}
					t.Logf("tested operation %s -> %s\n", op, err)
				}
			}
		}
	}

}
