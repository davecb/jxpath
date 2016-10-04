package pathExpr_test

import (
	"testing"
	"pathExpr"
	lexer "xml"
	"os"
	"syscall"
)

// The notation is odd, but taken from the go implementation.
func devNull() *os.File {
	return os.NewFile(uintptr(syscall.Stdin), "/dev/null")
}


// See if we can grab a world out of the right galaxy
func TestFindSuchThatBoundedBy(t *testing.T) {
	var tests = []struct {
		input  string
		expect string
	}{
		// find a galaxy where there is a world containg hello
		// {input: "<universe><galaxy><world>hello</world></galaxy></universe>", expect: `hello`},
		// {input: "<universe><galaxy><world>nada</world></galaxy><galaxy><world>hello</world><world/><galaxy></universe>", expect: `hello`},
		{input: "<universe><galaxy><world>nada</world></galaxy></universe>", expect: ``},
		{input: "<universe><galaxy></galaxy></universe>", expect: ``},

	}

	// suppress most i/o
	var x *os.File
	x, os.Stderr = os.Stderr, devNull()

	for i, test := range tests {
		path := pathExpr.NewPath(lexer.Lex(test.input))
		value := path.FindSuchThat("galaxy", "world", "hello", "universe").TextValue()

		if value != test.expect {
			t.Errorf("%d, mismatch: expected=%q got=%q",
				i, test.expect, value)
		}
	}
	os.Stderr = x
}

// Test the interpretation of path-expressions
func TestInterpreter(t *testing.T) {
	var tests = []struct {
		input  string
		expect string
	}{
		// Check for clean matches
		{input: `/universe/galaxy[world="hello"]`, expect: `hello who`},         // 0
		{input: `/universe/galaxy[world="hello"]/`, expect: `hello who`},        // 1
		{input: "/universe/galaxy[world=\"hello\"]/timelord", expect: `who`},    // 2
		{input: "/galaxy[world=\"hello\"]/timelord", expect: `who`},             // 3
		//  the next two lines test an unimplemented operation, so it's "fragile", in that
		// te test passes but the function will fail. FIXME implement it
		{input: `/universe[2]/galaxy[world="hello"]`, expect: `hello who`},      // 4
	}

	// suppress most i/o
	var x *os.File
	x, os.Stderr = os.Stderr, devNull()
	source := "<universe><galaxy><world>nada</world></galaxy>" +
		"<galaxy><world>hello</world><world/><timelord>who</timelord></galaxy></universe>"
	path := pathExpr.NewPath(lexer.Lex(source))

	for i, test := range tests {

		// interpret a path
		value := path.Interpreter(test.input) //, true)
		if value != test.expect {
			t.Errorf("%d, mismatch: expected=%q got=%q",
				i, test.expect, value)
		}
	}
	os.Stderr = x
}


// Test for warnings
func TestWarnings(t *testing.T) {

	var tests = []struct {
		input  string
		path   string
		expect string
	}{
		// fails on single elements
		// {input: `<empty/>`, path: `/empty`, expect: ``},

		// EndAt: warning: did not find the end of "empty" or even a begin "empty". Returning nil... []
		{input: `<empty />`, path: `/empty`, expect: ``},        // 1

		//  TextValue: warning, did not find non-blank text in the sub-path... []
		{input: `<begin><empty/></begin>`, path: `/empty`, expect: ``},

		// FindSuchThatBoundedBy, warning, did not find a galaxy such that world=hello.
		// ...  [{"BEGIN", "galaxy"} {"END", "galaxy"} {"BEGIN", "universe"} {"EOF", ""}]
		{input: `<universe><galaxy/><universe>`, path: `/universe/galaxy[world="hello"]`, expect: ``},
	}

	// to run this test, you need to either capture stderr and compare, or do it visually
	var x *os.File
	x, os.Stderr = os.Stderr, devNull()

	for i, test := range tests {
		path := pathExpr.NewPath(lexer.Lex(test.input))

		// interpret a path
		value := path.Interpreter(test.path)
		if value != test.expect {
			t.Errorf("%d, mismatch: expected=%q got=%q",
				i, test.expect, value)
		}
	}
	os.Stderr = x
}







