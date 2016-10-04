package token

// Type represents a lexical token's type from one of the kinds of
// tokenizers (xml, json and csv at the moment)
//
// To handle all three types, we need a named begin, a named end,
// a value, and just possibly a named attribute.
// Xml could demand all, and an elegant approach from Elliotte Rusty
// Howard is to treat attributes like a special case of a tag.
// Tags like <foo name=value /> become <begin foo><begin name>value
// <end name><end foo>.
// Json is more succinct, with {} as begin/end, but we still need
// a named end for doing "within".
// Finally, CSVs may be done with a vector of column names providing
// the names for the tags.
type Type int

// Token is what the jxpath interpreter interprets
type Token struct {
	Typ Type   // Type, such as BEGIN.
	Val string // Name, such as "universe".
}

// Pad is what append pads slices with.
const (
	PAD Type = iota
	ERROR
	EOF
	BEGIN
	VALUE
	END
)

var tokenTypes = [...]string {
	"PAD", // The type go will default to if the slice is auto-extended
	"ERROR",
	"EOF",
	"BEGIN",
	"VALUE",
	"END",
}

func (t Type) String() string {
	return "\"" + tokenTypes[t] + "\""
}

func (t Token) String() string {
	return "{" + t.Typ.String() + ", \"" + t.Val + "\"}"
}
