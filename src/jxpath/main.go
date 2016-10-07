package main

import (
	"token"
	xml_lexer "xml"		// lexer for xml
	json_lexer "json" 	// and for json
	"pathExpr"
	"trace"

	"fmt"
	"flag"
	"os"
	"io/ioutil"
	"strings"
)


/*
 * Parse an input string into a slice of tokens, as input for
 * one or more path expressions. Proof of concept for a general
 * path expression engine. Returns the result as a string on stdout,
 * 0 on success an 1 on error or warnings. Right now there are only
 * warnings.
 */
func main() {
	var source string
	var pathExpression string
	var inputType string
	var t trace.Trace
	var x, j, c, explain, tracing bool

	flag.BoolVar(&x, "xml", false, "parse xml input")
	flag.BoolVar(&j, "json", false, "parse json input")
	flag.BoolVar(&c, "csv", false, "parse csv input")
	flag.BoolVar(&explain, "explain", false, "explain what code to use")
	flag.BoolVar(&tracing, "trace", false, "trace in detail")

	flag.Parse();
	if x {
		if j || c {
			fmt.Fprint(os.Stderr, "more than one of -x, -j and -c called, -x taken\n")
			flag.Usage()
		}
		inputType = "xml"

	} else if j {
		if c {
			fmt.Fprint(os.Stderr, "more than one of -j and -c called, -j taken\n")
			flag.Usage()
		}
		inputType = "json"

	} else if c {
		inputType = "csv"
	}


	// Look for path expressions on the command-line
	// In future, allow one or more files|path-expressions
	if flag.NArg() == 0 {
		fmt.Fprint(os.Stderr, "Usage: jxpath path-expression*\n  Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Look on stdin for input data
	file := os.Stdin
	fi, err := file.Stat()
	if err != nil {
		// Stdin is broken?  Not much we can do.
		fmt.Println("os.Stdin failed to stat, halting", err)
		os.Exit(3)

	} else if fi.Size() > 0 {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin, %s, halting", err);
			os.Exit(3)
		}
		source = string(bytes)

	} else {
		// Report there wasn't anything to do
		flag.Usage()
		fmt.Fprint(os.Stderr, "No input was found on stdin, halting\n")
		os.Exit(1)
	}

	if tracing {
		t = trace.NewTrace(os.Stderr)
	} else {
		t = trace.NewTrace(ioutil.Discard)
	}

	// All the options are known, now use them
	defer t.Begin()()
	t.Printf("args=%s\n", flag.Args())

	if (inputType == "") {
		inputType = guessType(source, t)
		t.Printf("mime-type=%s\n", inputType)
	}

	var i int
	if inputType == "xml" {
		tokens := xml_lexer.Lex(source, t)
		for i, pathExpression = range flag.Args() {
			value := evaluate(tokens, pathExpression, explain, t)
			fmt.Printf("%d: path expression %q selected %q\n", i, pathExpression, value)
		}
	} else if inputType == "json" {
		tokens := json_lexer.Lex(source, t)
		for i, pathExpression = range flag.Args() {
			value := evaluate(tokens, pathExpression, explain, t)
			fmt.Printf("%d: path expression %q selected %q\n", i, pathExpression, value)
		}

	} else if inputType == "csv" {
		// and eventually if -c, csv files
		fmt.Fprint(os.Stderr, "Sorry, .csv isn't implemented yet.")
		return
	}
}



// evaluate applies the path expression to the tokenized inputs
func evaluate(tokens []token.Token, pathExpression string, explain bool, t trace.Trace) string {
	defer t.Begin(tokens, explain, t)()

	path := pathExpr.NewPath(tokens, t)
	value := path.Interpreter(pathExpression, t, explain)

	t.Printf("return path = %s\n", path)
	t.Printf("return value = %s\n", value)
	return value

}

// guessType guesses at the type of a file: xml and json are the only interesting ones,
// and just maybe .csv
func guessType(s string, t trace.Trace) string {
	defer t.Begin(s)()
	if strings.Contains(s, "<?xml") || strings.Contains(s, "</") || strings.Contains(s, "/>") {
		return "xml"
	} else if strings.Contains(s, ":") || strings.Contains(s, "{") {
		return "json"
	} else if strings.Contains(s, ",") {
		return "csv"
	}
	return "unguessed"
}
