package main

import (
	"token"
	"trace"
	xml_lexer "xml"
	json_lexer "json"

	"testing"
	"os"
	"io/ioutil"
	"syscall"
)

var xmlInput =
	"<universe>" +
		"<galaxy>" +
			"<world>" +
				"nada" +
			"</world>" +
		"</galaxy>" +
		"<galaxy>" +
			"<world>" +
				"earth" +
			"</world>" +
			"<world/>" +    // A nameless world.
			"<timelord>" +
				"who" + // Dr Who, to be precise.
			"</timelord>" +
		"</galaxy>" +
		"<timelord>"  +
			"master" +  // Trapped outside of space and time.
		"</timelord>" +
	"</universe>"

var jsonInput =
	`"universe":  {` +
		`galaxy: {` +
			`world: "nada",` +
		`},` +
		`galaxy: {` +
			`world: "earth",` +
			`world: "",` + // A nameless world
			`timelord: "who",` + // Dr Who, to be precise
		`},` +
		`timelord: "master",` + // Trapped outside of space and time.
	`}`


// Do the usual debugging here, not in main.go
func TestDebug(t *testing.T) {
	var tracer trace.Trace
	tracer = trace.New(os.Stderr, true) // use stderr to trace
	//tracer = trace.New(ioutil.Discard, true) // and this to not

	//var tokens []token.Token = xml_lexer.Lex(xmlInput, tracer) // xml
	var tokens = json_lexer.Lex(jsonInput, tracer) // json
	//var tokens []token.Token = csv_lexer.Lex(jsonInput, tracer) // csv
	var tests = []struct {
		expr    string
		expect  string
	} {
		// current debug cases
		{ expr: `universe/galaxy[world="earth"]/timelord`, expect: `who`},
	}

	tracer.Begin()()
	explain := false
	for i, test := range tests {
		value := evaluate(tokens, test.expr, explain, tracer)
		if value != test.expect {
			t.Errorf("%d: { expr:%q, expect:%q }, get %q\n",
				i, test.expr, test.expect, value)
		}
	}
}


// Benchmark the classic search
func BenchmarkDrWho(b *testing.B) {
	var tracer trace.Trace
	//tracer = trace.New(os.Stderr, true) // use stderr to trace
	tracer = trace.New(ioutil.Discard, true) // and this to not

	var tokens = xml_lexer.Lex(xmlInput, tracer) // xml
	//var tokens = json_lexer.Lex(jsonInput, tracer) // json
	//var tokens []token.Token = csv_lexer.Lex(jsonInput, tracer) // csv
	var tests = []struct {
		expr    string
		expect  string
	} {
		// current debug cases
		{ expr: `universe/galaxy[world="earth"]/timelord`, expect: `who`},
	}

	tracer.Begin()()
	explain := false
	for _, test := range tests {
		for i := 0; i < b.N; i++ {
			evaluate(tokens, test.expr, explain, tracer)
		}
	}
}

// Benchmark the whole suite, for the profiler
func BenchmarkEndToEnd(b *testing.B) {
	var tracer trace.Trace

	for i := 0; i < b.N; i++ {
		//tracer = trace.New(os.Stderr, true) // use stderr to trace
		tracer = trace.New(ioutil.Discard, true) // and this to not

		var tokens = xml_lexer.Lex(xmlInput, tracer) // xml
		//var tokens = json_lexer.Lex(jsonInput, tracer) // json
		//var tokens []token.Token = csv_lexer.Lex(jsonInput, tracer) // csv
		var tests = []struct {
			expr   string
			expect string
		}{
			// current debug cases
			{expr: `universe/galaxy[world="earth"]/timelord`, expect: `who`},
		}

		tracer.Begin()()
		explain := false
		for _, test := range tests {
			evaluate(tokens, test.expr, explain, tracer)
		}
	}
}

// Benchmark the lexer alone
func BenchmarkLexer(b *testing.B) {
	var tracer trace.Trace
	//tracer = trace.New(os.Stderr, true) // use stderr to trace
	tracer = trace.New(ioutil.Discard, true) // and this to not

	for i := 0; i < b.N; i++ {
		xml_lexer.Lex(xmlInput, tracer) // xml
		//json_lexer.Lex(jsonInput, tracer) // json
		//csv_lexer.Lex(jsonInput, tracer) // csv
	}
}


// Test the main-line as a function, these are the successes
func TestXmlPaths(t *testing.T) {
	var tracer trace.Trace   // use stderr to trace
	//tracer = trace.New(os.Stderr, true)
	tracer = trace.New(ioutil.Discard, true) // and this to not

	var tokens = xml_lexer.Lex(xmlInput, tracer)
	goodPathTests(t, tokens, tracer);
}

func TestJsonPaths(t *testing.T) {
	var tracer trace.Trace   // use stderr to trace
	//tracer = trace.New(os.Stderr, true)
	tracer = trace.New(ioutil.Discard, true) // and this to not

	var tokens = json_lexer.Lex(jsonInput, tracer)
	goodPathTests(t, tokens, tracer);
}

func goodPathTests(t *testing.T, tokens []token.Token, tracer trace.Trace)  {
	var tests = []struct {
		expr    string
		expect  string
	} {
		// regular success cases
		{ expr:"/world", expect:"nada", },
		{ expr:"//world", expect:"nada", },
		{ expr:"/galaxy/world", expect:"nada", },
		{ expr:"/universe/galaxy/world", expect:"nada", },
		{ expr: `/universe/galaxy[world="earth"]`, expect: `earth  who`},
		{ expr: `/universe/galaxy[world="earth"]/timelord`, expect: `who`},
		{ expr: `/galaxy[2]/timelord`, expect:`who` },
		{ expr: `/universe/galaxy[2]/timelord`, expect: `who`},
		{ expr: `/universe/timelord[2]`, expect: `master`},
	}

	tracer.Begin()()
	explain := false
	for i, test := range tests {
		value := evaluate(tokens, test.expr, explain, tracer)
		if value != test.expect {
			t.Errorf("%d: { expr:%q, expect:%q }, get %q\n",
				i, test.expr, test.expect, value)
		}
	}
}


// Test things that issue warnings
func TestXmlBadPaths(t *testing.T) {
	var tracer trace.Trace   // use stderr to trace
	//tracer = trace.New(os.Stderr, true)
	tracer = trace.New(ioutil.Discard, true) // and this to not

	var tokens = xml_lexer.Lex(xmlInput, tracer)
	badPathTests(t, tokens, tracer);
}

func TestJsonBadPaths(t *testing.T) {
	var tracer trace.Trace   // use stderr to trace
	//tracer = trace.New(os.Stderr, true)
	tracer = trace.New(ioutil.Discard, true) // and this to not

	var tokens = json_lexer.Lex(jsonInput, tracer)
	badPathTests(t, tokens, tracer);
}

func badPathTests(t *testing.T, tokens []token.Token, tracer trace.Trace)  {
	var tests = []struct {
		expr   string
		expect string
	}{
		// things that return blanks and generate 'warning, did not find'...)
		{ expr: `/galaxy[world="nada"]/timelord`, expect: ``}, //
		{ expr: `/galaxy[world="venus"]/timelord`, expect: ``},
		{ expr: `/galaxy[1]/timelord`, expect: ``},
		{ expr: `/universe/galaxy[3]/timelord`, expect: ``},
		{ expr: `/galaxy[3]/timelord`, expect: ``},
	}

	// to run a test on the _messages_ instead, you need to not shut of stderr
	var x *os.File
	x, os.Stderr = os.Stderr, devNull()

	explain := false
	for i, test := range tests {
		value := evaluate(tokens, test.expr, explain, tracer)
		if value != test.expect {
			t.Errorf("%d: { expr:%q, expect:%q }, get %q\n",
				i, test.expr, test.expect, value)
		}
	}
	os.Stderr = x
}


// devnull pretends to be a file, but does nothing
// The notation is odd, but taken from the go implementation.
func devNull() *os.File {
	return os.NewFile(uintptr(syscall.Stdin), "/dev/null")
}
