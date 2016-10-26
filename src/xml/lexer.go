// Package lexer -- lexer and parser for xml, one of trio of peer classes for xml, json and csv.
// FIXME this uses prefix, and peeks at the (exported) innards of lexer.Lexer
// changing to use a prefix in lexer
package lexer

import (
	"token"
	"lexer"
	"trace"

	"strings"
	"unicode"
)



// stateFn represents the state of the scanner
// as a function that returns the Next state.
type stateFn func(*lexer.Lexer) stateFn

var eof = -1

// Lex -- the entry point to the xml lexer
func Lex(Input string, tp trace.Trace) ([]token.Token) {
	Input = strings.TrimSpace(Input)
	l := lexer.New(Input, make(chan token.Token), tp)
	
	defer l.Begin()()

	go run(l) // Run the lexer, closes pipe

	// Simulate a parser, return only when done
	var slice []token.Token
	var tok token.Token

	for tok = range l.Pipe {
		//tok = <- l.Pipe
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
	var ch int
	var s string

	defer l.Begin()()
	// Process the first character
	ch = l.Next()
	if ch != '<' {
		l.Backup();
		l.Print("redirect to lexText\n")
		return lexText
	}

	// We have a <, do we have an </ or not?
	l.Printf("right now, start is at %.40q ...\n", l.Rest())
	l.Ignore()
	tokenTypeFound = token.BEGIN // Subject to change, though
	ch = l.Next()
	if ch == '/' {
		l.Ignore()
		tokenTypeFound = token.END
	} else {
	       l.Backup()
	}
	// postcondition: we hit < or </

	// Process the characters up to the end character
	for {
		ch = l.Next()
		//l.Printf("got %c\n", ch)

		//  handle />
		if ch == '/' {
			// Then we hit /> or a grammar error
			l.Backup()
			if len(l.Current()) > 0 {
				l.Emit(tokenTypeFound, l.Current())
				// emit an empty value
				l.Emit(token.VALUE, "")
				// and then end the type
				l.Emit(token.END, l.Current())
			}
			l.Next()
			l.Next()
			l.Ignore()
			// otherwise there was nothing there or we just
			// processed some attributes
			return lexText
		}

		// handle >
		if ch == '>' {
			l.Backup()
			if len(l.Current()) > 0 {
				l.Emit(tokenTypeFound, l.Current())
			}
			l.Next()
			l.Ignore()
			return lexText
		}

		// handle whitespace, meaning do zero or more
		// attributes, don't eat > or />
		if unicode.IsSpace(rune(ch)) {
			// Emit the begin or end here
			l.Backup()
			l.Emit(tokenTypeFound, l.Current())
			l.Ignore()
			lexAttributes(l)
		}

		// Bail on eof
		if ch == eof {
			l.Backup()
			break
		}
	}
	// do parser.Token end and eof if unclosed
	l.Print("Emiting output\n")
	s =  l.Current()
	if len(s) > 0 {
		l.Emit(tokenTypeFound, s)
	}
	l.Emit(token.EOF, "")
	return nil
}

// lexAttributes starts a subloop getting name=value pairs until >
func lexAttributes(l *lexer.Lexer) {
	defer l.Begin()()

	for state := lexOneAttr; state != nil; {
		state = state(l)
	}
}

// lexOneAttr gets one name=value pair
// Restricted to x=y with no spaces around the =
func lexOneAttr(l *lexer.Lexer) stateFn {
	var ch int
	defer l.Begin()()

	l.Next() // skip whitespace
	l.Ignore()
	for {
		ch = l.Next()
		l.Printf("got %c\n", ch)

		if ch == '/' || ch == '>' || ch == eof {
			// Then we hit /> or a grammar error
			l.Backup()
			if len(l.Current()) > 0 {
				l.Emit(token.VALUE, l.Current())
				l.Emit(token.END, l.Pop())
			}
			return nil
		}

		if unicode.IsSpace(rune(ch)) {
			// end of the attribute
			l.Backup()
			if len(l.Current()) > 0 {
				l.Emit(token.VALUE, l.Current())
				l.Emit(token.END, l.Pop())
			}
			return lexOneAttr
		}

		if ch == '=' {
			// we have the name, save it and start the value
			l.Backup()
			l.Push(l.Current())
			l.Emit(token.BEGIN, l.Current())
			l.Next()
			l.Ignore()
		}

		if ch == '"' {
			l.Backup()
			s := l.AcceptQstring()
			l.Emit(token.VALUE, s)
			l.Emit(token.END, l.Pop())
		}

	}
}


// Lex text as a VALUE
func lexText(l *lexer.Lexer) stateFn {
	var ch int

	defer l.Begin()()
	l.Printf("Input=%.40q ...\n", l.Rest())
	if l.Rest() == "" {
		l.Print("Emitting EOF, returning nil\n")
		l.Emit(token.EOF, "")
		return nil      // Stop the run loop.
	}
	for {
		ch = l.Next()
		if ch == '<' {
			l.Backup()
			s := l.Current()
			if len(s) > 0 {
				l.Emit(token.VALUE, s)
			}
			l.Print("redirect to lexTag\n")
			return lexTag // Next state.
		}
		if ch == eof {
			l.Backup()
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

