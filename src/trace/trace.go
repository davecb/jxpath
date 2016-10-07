/*
 * A go version of the Solaris apptrace logger.
 * It will get better as I learn more go. Right now
 * it likes things sequentialized, which is easy 
 * in lexer->parser schemes. Georg Nikodym did
 * the concurrent version on Solaris. Mine will
 * probably need a different log stream passed in
 * for each distinct tracer.
 */

// Package trace implements a "log in context"
package trace

import (
	"log"
	"io"
	"runtime"
	"fmt"
	"io/ioutil"
)

// Trace -- the "apptrace(1)" view of a logger
type Trace interface {
	Begin(...interface{}) func()
	Printf(format string, v ...interface{})
	Print(format string)
}

// RealTrace really traces: Fake doesn't
type RealTrace struct {
	pipe  chan message
	log   *log.Logger
}

// message is what is passed through the pipe
type message struct {
	r rune    	// an indicator for indent & outdent
	s string     	// the string to indent
}

// NewTrace -- create one or more type of tracing program
func NewTrace(fp io.Writer) Trace {
	var real RealTrace
	var fake FakeTrace
	if fp == ioutil.Discard {
		// do discarding less expensively
		return &fake
	}
	real.log = log.New(fp, "", 0)

	real.log.SetFlags(0)
	real.pipe = make(chan message)

	go real.backend()
	return &real
}

// Begin a function and return an at-end function for defer.
func (t RealTrace) Begin(args ...interface{}) func() {
	var s string

	pc, _, _, _ := runtime.Caller(2)
	name := runtime.FuncForPC(pc).Name()

	seperator := ""
	for _, arg := range args {
		s += fmt.Sprintf("%s%v", seperator, arg)
		seperator = ", "
	}
	t.pipe <- message{r: '>', s: fmt.Sprintf("%s(%s)", name, s)}

	return func() {
		t.pipe <- message{r: '<', s: fmt.Sprintf("%s", name)}
	}
}

// Printf -- write indented to the trace stream
func (t RealTrace) Printf(format string, v ...interface{}) {
	t.pipe <- message{r: '|', s: fmt.Sprintf("%s", fmt.Sprintf(format, v...))}
}

// Print -- write indented to the trace stream
func (t RealTrace) Print(format string) {
	t.pipe <- message{r: '|', s: fmt.Sprintf("|%s", format)}
}


// backend does indentation and padding, as per the first param
func (t RealTrace) backend() {
	var indent = -1
	var direction string
	var msg message

	for {
		msg = <-t.pipe
		switch (msg.r) {
		case '>':
			direction = "begin "
			indent++
		case '<':
			direction = "end "
		case '|':
			direction = ""
		}
		t.log.Printf("%s%s%s", t.pad(indent), direction, msg.s)
		if msg.r == '<' {
			indent--
		}
	}
}

// Pad the string to as much as 72 places.
func (t *RealTrace) pad(depth int) string {

	const padding string = "   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |    "
	                      //-123456780-123456789-123456789-123456789-123456789-123456789-123456789-12
	                      //          1         2         3         3         5         6         7

	offset :=  len(padding)-(depth*4)
	if offset <= 0 {
		// we just stop incrementing.  This just indicates a DEEP nesting
		offset = 0
	} else if offset > len(padding) {
		// we just stop decrementing.  This indicates an extra end
		//fmt.Fprintf(os.Stderr,"programmer error, offset %d > length %d\n", offset, len(padding))
		offset = len(padding)
	}
	return padding[offset:]
}


// FakeTrace does very little
type FakeTrace struct {}

// Begin and end a per-function trace, by doing nothing twice
func (* FakeTrace) Begin(args ...interface{}) func() {
	return func() {}
}

// Printf -- write nothing
func (t FakeTrace) Printf(format string, v ...interface{}) {
	 //
}

// Print -- write nothing
func (t FakeTrace) Print(format string) {
	//
}
