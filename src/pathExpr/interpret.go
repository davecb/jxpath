package pathExpr

import (
	"trace"

	"strings"
	"fmt"
	"os"
	"unicode"
	"unicode/utf8"
	"strconv"
)
var t trace.Trace

// Interpreter reads a string like /universe/world or /match[opponent="fred"]
// and return either a value or an explanation. For learning purposes, this
// is a non-lexing string-walk.
func (p Path) Interpreter(expression string, t trace.Trace, returnExplanation ...bool) string {
	var elementName, expressionName, expressionValue string
	var previousElement = ""
	var parseThusFar = "/"  // for error reporting
	var i int
	var explanation = "path := pathExpr.NewPath(lexer.Lex(input)); value := path"
	var explain = false
	defer t.Begin(expression)()

	for _, x := range returnExplanation {
	     explain = x
	}
	// if a value is not passed in, explain will remain false

	// Trim any trailing slashes and spaces
	var length = len(expression)
	if strings.HasPrefix(expression[length-1:length], "/") {
		// remove the /
		expression = expression[:length-1]
	}
	for i = 0; i < len(expression); i++ {
		//t.Printf("i=%d, %s\n", i, expression[i:])
		if strings.HasPrefix(expression[i:], "/") {
			// it's a new component
			if i != 0 {
				// record all non-empty steps for diagostics
				parseThusFar += recordParse(elementName, expressionName, expressionValue, previousElement)
				parseThusFar += "/"

				// record how we'd do the step
				explanation += recordStep(elementName, expressionName, expressionValue, previousElement)

				// and then execute them
				p = doStep(p, elementName, expressionName, expressionValue, previousElement)
			}
			previousElement = elementName
			elementName, expressionName, expressionValue = "", "", ""

		} else if strings.HasPrefix(expression[i:], "[") {
			// it's an expression
			var count int
			count, expressionName, expressionValue = p.expressionEval(expression[i+1:])
			i += count +1

		} else {
			// it's part of the component's name
			elementName += expression[i:i + 1]
		}
	}
	// Add the current component to the parse
	parseThusFar +=  recordParse(elementName, expressionName, expressionValue,
		previousElement)

	// execute the last step
	p = doStep(p, elementName, expressionName, expressionValue, previousElement)

	// get the value of the last element
	result := p.TextValue()

	// explain
	explanation += recordStep(elementName, expressionName, expressionValue,
		previousElement) + ".TextValue()"

	// and then report
	t.Printf("parse=%s\n", parseThusFar)
	t.Printf("explanation=%s\n", explanation)
	if explain {
		// Write it to stdout for the engineer to copy
		fmt.Fprintf(os.Stderr, "explanation: %s\n", explanation)
	}
	return result
}

func (p Path) expressionEval(s string) (int, string, string) {
	var i int
	var name, value string
	var inVal = false
	defer t.Begin(s)()

	r, _:= utf8.DecodeRuneInString(s[0:1])
	if unicode.IsDigit(r) {
		// it's a numbed selection, a special case
		for i = 0; i < len(s); i++ {
			//t.Printf("i=%d, %s\n", i, s[i:])
			if strings.HasPrefix(s[i:], "]") {
				// we're done
				break
			}
			if r, _:= utf8.DecodeRuneInString(s[0:1]); unicode.IsDigit(r) {
				value += s[i:i + 1]
			} else {
				fmt.Fprintf(os.Stderr, "Warning, %q is not a number, so " +
					"this selection expressions is invalid, and the " +
					"result may be wrong due to this error in the input, %s\n",
					s[0:1], s)
				// ignore value
			}
		}
		name = ""
		t.Printf("i = %d, name=%s, value=%s", i, name, value)
		return i, name, value
	}

	// Otherwise it's a name=value expression
	for i = 0; i < len(s); i++ {
		//ft.Printf("i=%d, %s\n", i, s[i:])
		if strings.HasPrefix(s[i:], "=") {
			inVal = true

		} else if strings.HasPrefix(s[i:], "]") {
			// we're done
			break

		} else if strings.HasPrefix(s[i:], "\"") {
			// do nothing and thus skip the quote

		} else {
			if inVal == true {
				// if text, it's the component's value
				value += s[i:i + 1]
			} else {
				// or name
				name += s[i:i + 1]
			}
		}
	}
	return i, name, value
}

func recordStep(componentName, expressionName, expressionValue, previousComponent string) string {
	if componentName == "" {
		// The expression contained a trailing slash, drop it
		return ""
	}

	// componentName[expressionName=expressionValue]
	if expressionName != "" {
		return `.FindSuchThat("` + componentName + `", "` +
			expressionName + `", "` + expressionValue +`")`
	}

	// componentName[2]
	if expressionValue != "" {
		return `.FindNth("` + componentName + `", "` +
			  expressionValue +`")`
	}

	// fcomponentName
	return `.FindFirst("` + componentName + `")`
}

// do a step
func doStep(path Path, componentName, expressionName, expressionValue, previousComponent string) Path {
	if componentName == "" {
		// The expression contained a trailing slash, ignore it (old?)
		return path // unchanged
	}

	// componentName[expressionName=expressionValue]
	if expressionName != "" {
		path = path.FindSuchThat(componentName, expressionName,
			expressionValue)
		//t.Printf("ran FindSuchThat(%s, %s, %s), got %s\n",
		//	componentName, expressionName, expressionValue, path)
		return path
	}

	// componentName[2]
	if expressionValue != "" {
		i, _ := strconv.Atoi(expressionValue)
		path = path.FindNth(componentName, i)
		t.Printf("ran FindNth(%s, %d), got %s\n",
			componentName, i, path)
		return path
	}

	//componentName
	path = path.FindFirst(componentName)
	//t.Printf("ran .FindFirst(%s), got %s\n", componentName, path)
	return path
}


func recordParse(componentName, expressionName, expressionValue, previousComponent string) string {
	var parse = componentName

	if componentName == "" {
		return ""
	}

	if expressionName != "" {
		parse += "[" + expressionName + `="` + expressionValue + `"]`
		return parse
	}

	if expressionValue != "" {
		parse += "[" + expressionValue + "]"
		return parse
	}

	return parse
}
