package operations

import (
	"testing"
)

var wrong_ops = [...]string{"1.755+", "1.755+)", "(1.755+", "(1.755+)", "1.755+]", "[1.755+", "[1.755+]", "1.755+}", "{1.755+", "{1.755+}", "(1.755+}", "[1.755+}"}

func TestWrongParse(t *testing.T) {
	for _, op := range wrong_ops {
		_, err := Parse_string(op)
		if err == nil {
			t.Errorf("%s did not errored out", op)
		}
	}

}
