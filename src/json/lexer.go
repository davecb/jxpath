// Package json -- lexer and parser for json, one of trio of peer classes for xml, json and csv.
package json

import (
	"token"
	"trace"
	"lexer"

	"fmt"
	"strings"
	"unicode"
)

//var t trace.Trace

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer.Lexer) stateFn

// Type jLex composes a low-level Lexer into this one
type jLex struct {
     *lexer.Lexer     // the lower-level lexer
     trace.Trace // compose in Begin. Print, Printf
}

const eof = -1  	// see note in lexer re is this good or not

// Lex is the entry point to the json lexer
func Lex(input string, tp trace.Trace) ([]token.Token) {
	
	var slice  = make([]token.Token, 0)
	input = strings.TrimSpace(input)
	l := lexer.New(input, make(chan token.Token), tp)
	defer l.Begin()()

	go run(l) // closes pipe
	slice = parse(l, slice)
	l.Printf("returning %s\n", slice)
	return slice
}

// parse accepts the lexemes from run and returns when it has them all
// Right now it's the null parser. Not shared, as it's not always going
// to be null.
func parse(l *lexer.Lexer, slice []token.Token) []token.Token {
	var tok token.Token

	//defer l.Begin()()
	for {
		tok = <- l.Pipe
		slice = append(slice, tok)
		//l.Printf("parse: appending %v to the slice\n", tok)
		if tok.Typ == token.EOF || tok.Typ == token.ERROR || tok.Typ == token.PAD {
			l.Printf("at end, token = %s", tok)
			break
		}
	}
	//l.Printf("parse: at end of parse loop, slice = %v\n", slice)
	return slice
}

// Run lexes the Input by executing state functions until
// the state is nil, then closes its output
func run(l *lexer.Lexer) {
	defer l.Begin()()

	for state := lexUnnamedBegin; state != nil; {
		state = state(l)
	}
	// FIXME is this reached? Yes, I had a *t vs t bug!
	l.Print("At end of l.run(), is this recorded?")
	close(l.Pipe)
}


// lexUnnamedBegin recognizes a naked "{" at the beginning of
// a json expression.
func lexUnnamedBegin(l *lexer.Lexer) stateFn {
	defer l.Begin()()

	l.Printf("starting with %.40q ...\n",l.Rest())
	l.SkipOver()
	if strings.HasPrefix(l.Rest(), "{") {
		l.Next()
		l.Emit(token.BEGIN, "")
		l.Push("<unnamed>")
		return lexName
	}
	// otherwise start looking for a name
	return lexName
}

// lexJson lexes a json name, ending in a colon:
func lexName(l *lexer.Lexer) stateFn {
	var cantidateName string
	defer l.Begin()()

	l.Printf("starting with %.40q ...\n",l.Rest())
	l.SkipOver()
	// Expect },  letters, qstring, or eof
	var nextc = l.Next()
	if nextc == '}' {
		// We hit the end of a block
		name := l.Pop()
		l.Printf("found a }, ending %s\n", name)
		l.Emit(token.END, name)
		return lexName

	} else 	if nextc == '"' {
		// Found a double-quote, it's a qstring
		l.Ignore()
		cantidateName = l.AcceptVariableName()
		l.Next()
		l.Ignore()
		l.Printf("got quoted text `%s`\n", cantidateName)

	} else if unicode.IsLetter(rune(nextc)) {
		// Found an ordinary unquoted name
		l.Backup()
		cantidateName = l.AcceptVariableName()
		//cantidateName = l.Current()
		l.Printf("got text `%s`\n", l.Current())

	} else if (nextc == eof) {
		l.Print("got eof")
		l.Emit(token.EOF, "")
		return nil

	} else {
		l.Emit(token.ERROR, fmt.Sprintf("Needed a string or qstring to make up a name, got %c (%v)" ,nextc, nextc))
		return nil
	}
	// Postcondtion: we have a name, candidate for a <BEGIN name>

	// expect a colon
	l.Printf("continuing with %.40q ...\n", l.Rest())
	l.SkipOver()
	nextc = l.Next()
	if nextc == ':' {
		l.Printf("got a name, %s\n", cantidateName)
		// success, Push name here
		l.Push(cantidateName)
		l.Emit(token.BEGIN, cantidateName)
		return lexValue;

	} else if (nextc == eof) {
		l.Emit(token.EOF, "")
		return nil
	} else {
		l.Emit(token.ERROR, fmt.Sprintf("Needed a : to complete a name, got %c (%v)", nextc, nextc))
		return nil
	}
}

// lexValue recognizes begin-block ("{") and qstring values
func lexValue(l *lexer.Lexer) stateFn {
	var cantidateValue string
	defer l.Begin()()
	l.Printf("starting with %.40q ...\n",l.Rest())
	l.SkipOver()

	// Expect {, qstring, or eof
	var nextc = l.Next()
	if nextc == '"' {
		// Found a double-quote, it's a qstring
		l.Backup()
		cantidateValue = l.AcceptQstring()
		l.Emit(token.VALUE, cantidateValue)
		name := l.Pop()
		l.Emit(token.END, name)
		return lexName

	} else if (nextc == '{') {
		// we've started a sequence of name:value statements, separated with commas
		// that are the contents of the last BENIN <name>.Start gobbling with lexName/lexValue
		l.Ignore()
		return lexName

	} else if (nextc == eof) {
		// we've fallen off the end unexpectedly
		l.Emit(token.EOF, "")
		return nil
	}
	l.Emit(token.ERROR, fmt.Sprintf("Needed a value to complete a name:value pair, got %c (%v)", nextc, nextc))
	return nil
}


/*
 * Things to test
 */
// with and without initial {, <?xml?>
// with _ in names
// with blanks in values
// with qstring names
// with illegal first char  in names
// with eof in whitespace, names and values
// wilh ill-formed qstring, s = ` "universe: `, returns EOF and s
// `{ "universe": "quantity one" }
// an array of galaxies
//  an eof-in-qstring test to validate
// a \" in a qstring
