// Package lexer -- lexer and parser for xml, one of trio of peer classes for xml, json and csv.
// FIXME this uses prefix, and peeks at the (exported) innards of lexer.Lexer
// changing to use a prefix in lexer
package lexer

import (
	"token"
	"lexer"
	"trace"

	"strings"
)



// stateFn represents the state of the scanner
// as a function that returns the Next state.
type stateFn func(*lexer.Lexer) stateFn

var eof = -1

// Lex -- the entry point to the xml lexer
func Lex(Input string, tp trace.Trace) ([]token.Token) {
	Input = strings.TrimSpace(Input)
	l := lexer.New(Input, make(chan token.Token), tp)
	
	defer l.Begin(Input)()

	go run(l) // Run the lexer, closes pipe

	// Simulate a parser, send done when done
	var slice []token.Token
	var tok token.Token
	for {
		tok = <- l.Pipe
		slice = append(slice, tok)
		if tok.Typ == token.EOF {
			break
		}
	}
	
	return slice
}

// Run lexes the Input by executing state functions until
// the state is nil.
func  run(l * lexer.Lexer) {
	defer l.Begin(l)()
	for state := lexTag; state != nil; {
		state = state(l)
	}
	close(l.Pipe)
}



// lexTag lexes an xml tag
func lexTag(l *lexer.Lexer) stateFn {
	var tokenTypeFound token.Type

	defer l.Begin(l.Rest())()
	// Process the first character
	if ! l.HasPrefix("<") {
		l.Backup();
		l.Print("redirect to lexText\n")
		return lexText
	}

	// We have a <, do we have an </ or not?
	l.Printf("right now, Start is at %s\n", l.Rest())
	l.Next() // skip past the <
	l.Ignore()
	l.Printf("after ignore, Start is at %s\n", l.Rest())
	tokenTypeFound = token.BEGIN // Subject to change, though

	if l.Next() == '/' {
		l.Ignore()
		tokenTypeFound = token.END
	} else {
	       l.Backup()
	}

	// Process the remaining characters
	for {
		// handle <foo/>
		if l.HasPrefix("/>") {
			l.Emit(tokenTypeFound, l.Current())
			l.Emit(token.VALUE, "")
			l.Emit(token.END,l.Current())
			l.Next()
			l.Next()
			l.Ignore()
			return lexText
		}

		// consume, discarding including tag-end character
		if l.HasPrefix(">") {
			l.Emit(tokenTypeFound, l.Current())
			l.Next()
			l.Ignore()
			return lexText
		}

		// Bail on eof
		if l.Next() == eof {
			break
		}
	}
	// do parser.Token end and eof if unclosed
	l.Print("Emiting output\n")
	s :=  l.Current()
	if len(s) > 0 {
		l.Emit(token.BEGIN, s)
	}
	l.Emit(token.EOF, "")
	return nil
}


// Lex text as a VALUE
func lexText(l *lexer.Lexer) stateFn {

	defer l.Begin(l.Rest())()
	l.Printf("Input=%q\n", l.Rest())
	if l.Rest() == "" {
		l.Print("Emitting EOF, returning nil\n")
		l.Emit(token.EOF, "")
		return nil      // Stop the run loop.
	}
	for {
		if l.HasPrefix("<") {
			s := l.Current()
			if len(s) > 0 {
				l.Emit(token.VALUE, s)
			}
			l.Print("redirect to lexTag\n")
			return lexTag // Next state.
		}
		if l.Next() == eof {
			break
		}
	}
	s:= l.Current()
	if len(s) > 0 {
		l.Print("Emitting output\n")
		l.Emit(token.VALUE, l.Current())
	}
	l.Emit(token.EOF, "")
	return nil
}

