// Package lexer is a collection of methods for the xml and json lexers.
// Almost everything in it is a public shared implementation.
package lexer

import (
	"token"
	"trace"

	"unicode/utf8"
	"unicode"
	"strings"
	"fmt"
)

const eof = -1  // is this a good idea or unneeded complexity?
		// FIXME figure out of empty strings are better than eofs

//var t  trace.Trace


// Lexer is the underlying data structure for the two language-specific lexers
type Lexer struct {
	input string           // the string being scanned.
	start int              // start position of this item.
	pos   int              // current position in the input.
	width int              // width of last rune read from input.
	stack []string         // for begin-end matching
	Pipe  chan token.Token // channel of parser.Tokens.
	trace.Trace            // a composed-in tracer
}

// New creates a lexer struct
func New(input string, pipe chan token.Token, tp trace.Trace) (*Lexer) {
	var l = Lexer{input, 0, 0, 0, nil, pipe, tp}
	return &l
}

// String displays a minimal view of the Lexer FIXME
func (l *Lexer) String() string {
	return  fmt.Sprintf(
		"{ input: %v\n" +
		"  start: %d\n" +
		"  pos: %d\n" +
		"  width: %d\n" +
		"  stack: %v\n" +
		"  Pipe: %v\n"  +"}",
		l.input, l.start, l.pos, l.width,
		l.stack, l.Pipe)
}

// AcceptQstring parses a quoted string
func (l *Lexer) AcceptQstring() string {
	var nextc int
	var s string

	defer l.Begin()()
	l.Printf("starting with %.40q ....\n",l.Rest())
	l.Next()  // strip off "
	l.Ignore()
	for {
		if nextc = l.Next(); nextc == '"' || nextc == eof {
			//l.Printf("rejected %q\n", nextc)
			break
		}
		if nextc == '\\' {
			// accept this and the next character blindly
			l.Next()
		}
		//l.Printf("accepted %q\n", nextc)
	}
	l.Backup()
	s = l.Current()
	l.Printf("returning %q\n", s)
	l.Next()
	l.Ignore()
	return s
}

// AcceptVariableName parses a variable-name
func (l *Lexer) AcceptVariableName() string {
	var nextc int

	defer l.Begin()()
	for {
		nextc = l.Next()
		if !unicode.IsLetter(rune(nextc)) &&
			!unicode.IsNumber(rune(nextc)) &&
			nextc != '_' {
			//l.Printf("rejected %q\n", nextc)
			break
		}
		//l.Printf("accepted %q\n", nextc)
	}
	if (nextc == eof) {
		l.Print("unexpected eof")
		l.Emit(token.EOF, l.Current())
		return ""
	}
	l.Backup()
	return l.Current()
}

// Emit passes an item to the parser via the pipe.
func (l *Lexer) Emit(tt token.Type, s string) {
	value :=  token.Token{Typ: tt, Val:s}
	l.Pipe <- value
	l.start = l.pos // advance to pos
}

/*
 * Functions for traversing characters (runes)
 */

// Current returns the string we've collected to date
func (l *Lexer) Current() string {
	return l.input[l.start:l.pos]
}

// Rest returns the remaining characters to be lexed, after pos.
func (l *Lexer) Rest() string {
	return  l.input[l.pos:]
}

// Next returns the next rune, as an int
// FIXME why not a rune? eof == -1, that's why...
func (l *Lexer) Next() int {
	var r rune
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width =
		utf8.DecodeRuneInString(l.Rest())
	l.pos += l.width
	return int(r)
}


// Ignore skips over the pending input before this point.
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// Backup steps back one rune.
// Can be called only once per call of next.
func (l *Lexer) Backup() {
	l.pos -= l.width
}

// SkipOver skips over whitespace and commas, ignoring them.
// FIXME take out commas later, or make into a ...parameter
//
func (l *Lexer) SkipOver() {
	defer l.Begin()()
	for {
		nextc := l.Next()
		if unicode.IsSpace(rune(nextc)) {
			l.Printf("skipped whitespace %q\n", nextc)
		} else if nextc == ',' {
			l.Print("skipped comma\n")
		} else {
			break // something else
		}
	}
	l.Backup() // we're at a non-whitespace, back up one
	l.Ignore() // ignore the whitespace seen
}


/*
 * A stack of names, for begin/end matching
 */

// Push pushes a <BEGIN name>'s name on the stack
func (l *Lexer) Push(name string ) {
	defer l.Begin(name)()

	l.stack = append(l.stack, name)
	l.Printf("pushing onto %v\n", l.stack)
}

// Pop pops a <BEGIN>' name off for an <END name>
func (l *Lexer) Pop() string {
	defer l.Begin()()

	l.Printf("popping from %v\n", l.stack)
	length := len(l.stack)
	if length < 1 {
		return "STACK UNDERFLOW"
	}
	value :=  l.stack[length-1]
	l.stack = l.stack[:length-1]
	return value
}


// HasPrefix looks for a string without advancing. Used only
// in xml parse. Json uses pushback instead of lookahead.
func (l *Lexer) HasPrefix(s string) bool {
	defer l.Begin()()

	return strings.HasPrefix(l.input[l.pos:], s)
}
