/*
 * A go version of the Solaris apptrace logger.
 * It will get better as I learn more go. Right now
 * it likes things sequentialized, which is easy 
 * in lexer->parser schemes. Georg Nikodym did
 * the concurrent version on Solaris.
 */

package trace

import (
	"log"
	"io"
	"runtime"
	"fmt"
	"io/ioutil"
)

// Trace -- the "apptrace(1)" view of a logger
type Trace struct {
	log *log.Logger
	depth int
}

// Pad the string to as much as 72 places.
func (t *Trace) pad() string {
	const padding string = "   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |    "
	                     //-123456780-123456789-123456789-123456789-123456789-123456789-123456789-12
	                     //          1         2         3         3         5         6         7
	offset :=  len(padding)-(t.depth*4)
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

// NewTrace -- create a trace pointer
func NewTrace(fp io.Writer) *Trace {
	var t Trace
	if fp == ioutil.Discard {
		t.depth = -1
		return &t
	}
	t.log = log.New(fp, "", 0)
	t.log.SetFlags(0)
	t.depth = 0
	return &t
}

// Begin a function and return an at-end function for defer.
func (t *Trace) Begin(args ...interface{}) func() {
	var s string

	if t.depth == -1 {
		return func() { }
	}
	pc, _, _, _ := runtime.Caller(1)
	name := runtime.FuncForPC(pc).Name()

	seperator := ""
	for _,it := range args {
		s += fmt.Sprintf("%s%v", seperator, it)
		seperator = ", "
	}
	t.log.Printf("%sbegin %s(%s)\n", t.pad(), name, s)
	t.depth++

	return func() {
		t.depth--;
		t.log.Printf("%send %s\n", t.pad(), name)
	}
}

// Printf -- write indented to the trace stream
func (t *Trace) Printf(format string, v ...interface{}) {
	if t.depth == -1 {
		return
	}
	t.log.Printf("%s%s", t.pad(), fmt.Sprintf(format, v...))
}

// Print -- write indented to the trace stream
func (t *Trace) Print(format string) {
	if t.depth == -1 {
		return
	}
	t.log.Printf("%s%s", t.pad(), format)
}
