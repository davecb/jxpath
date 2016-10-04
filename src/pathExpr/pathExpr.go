/*
 * The primitives for a path engine
 */

package pathExpr

import (
	"trace"
	"token"

	"fmt"
	"strings"
	"os"
	"io/ioutil"
)

/*
 * Path expressions, interpreter and implementations
 */

// Path is a set of tokens to traverse
type Path []token.Token

var t *trace.Trace
var traceFp = ioutil.Discard
var warnings = 0  // Serially reusable, courtesy of this.
//var errors = 0  / future

// NewPath creates a new path from a slice of Tokens
func NewPath(input []token.Token, trace *trace.Trace) Path {
	t = trace
	defer t.Begin(input)()
	return input
}

// TextValue returns the text value within a bounded path. Do not use on a
// path with lots of values, you'll get all the texts
func (p Path) TextValue() string {
	var s string

	defer t.Begin(p)()
	for _, t := range p {
		if t.Typ == token.VALUE {
			s += strings.TrimSpace(t.Val) + " "
			// FIXME this could throw false positives due to <TEXT "\n"> tokens
			// the parser goroutine needs to squeeze them out
			//if i > 0 {
			//	fmt.Fprintf(os.Stderr, "TextValue: warning, more than one " +
			//		"text value in the specified path. The result may be " +
			//		"legitimately multiple, but it can also be wrong due " +
			//		"to an error in the input, %v\n", p)
			//	warnings++
			//}
		}
	}
	s = strings.TrimSpace(s)
	if (len(s) == 0 ) {
		fmt.Fprintf(os.Stderr, "TextValue: warning, did not find " +
			"non-blank text in the specified path. The result may be " +
			"legitimately blank, but it can also be wrong due " +
			"to an error in the input, %v\n", p)
		warnings++
	}
	return s
}

// Warnings returns the number of warnings made
func (p Path) Warnings() int {
	return warnings
}

// FindFirst finds the first instance of a sub-path within the global path.
func (p Path) FindFirst(target string) Path {

	defer t.Begin(target, p)()
	var beginning int
	// traverse input to target, return there to the end
	for  i := range p {
		//t.Printf("p[%d]=%v\n", i, p[i:i+1])
		if (p[i].Val == target && p[i].Typ == token.BEGIN) {
			t.Printf("begin is p[%d]=%v\n",
				i, p[i:i+1])
			beginning = i+1
		}
		if (p[i].Val == target && p[i].Typ == token.END) {
			if beginning == 0 {
				// we found an end first, skip over it
				// as these are a normal case in FindNext
				continue
			}
			t.Printf("end is p[%d]=%s\n",
				i, p[i:i+1])
			return p[beginning:i]
		}
	}
	// no end, return nil
	return nil
}


// FindNext finds the next element after the end of the previous one
func (p Path) FindNext(target string) Path {
	var q Path

	defer t.Begin(target, p)()
	if cap(p) > len(p) {
		// we can skip path the end of p?
		q = p[len(p):cap(p)]
	} else {
		q = nil
	}

	return q.FindFirst(target)
}

// FindSuchThat loops through elements, looking for tokenName == desiredValue and
// then return a subslice containing all of the matching elements.
func (p Path) FindSuchThat(element, tokenName, desiredValue string) Path {
	var q = p

	defer t.Begin(element, tokenName, desiredValue,p)()
	t.Printf("input=%s\n", p)

	// advance to the beginning of the first galaxy, at BEGIN element
	// search through the galaxies for tokenName == desiredValue
	for q = q.FindFirst(element); q != nil ; q = q.FindNext(element) {
		t.Printf("q=%s\n", q)

		subPath := q.FindFirst(tokenName)
		t.Printf("subpath=%s\n", subPath)
		if subPath.TextValue() == desiredValue {
			// return the beginning of the subpath q if we matched
			return q
		}
	}
	// didn't find it at all, which could be legit
	fmt.Fprintf(os.Stderr, "FindSuchThat: warning, did not find " +
		"a %s such that %s=%q. It may legitimately not exist, but " +
		"it can also be wrong due to an error in the input, %v\n",
		element, tokenName, desiredValue, p)
	warnings++
	return nil
}

// FindNth loops through n copies of element
func (p Path) FindNth(element string, n int) Path {
    	var q = p

	defer t.Begin(element, n, p)()
	i := 1 // for the findFirst
	for q = q.FindFirst(element); i < n && q != nil ; q = q.FindNext(element) {
		i++
	}
	return q
}



/*
 * for doc.go:
 * if the slice gets extended, the passings will be with "PAD"s, like
 * ... {"END", "universe"} {"EOF", ""} {"PAD", ""} {"PAD", ""} {"PAD", ""}
 * because they're zero-padded, and I assigned 0 to be PAD
 */
