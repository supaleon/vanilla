package token

import (
	"strconv"
	"unicode"
)

// Token is the set of lexical tokens of the Vanilla component programming language.
type Token uint8

const (
	ErrorToken   Token = iota // Special token: syntax error.
	EOFToken                  // Special token: end of a file.
	CommentToken              // <!--x-->
	DoctypeToken              // <!DOCTYPE x>
	CDATAToken                // <![CDATA[section]]>
	TextToken                 // abc
	SpaceToken                // ' '

	keywordBegin // Start of keyword tokens
	IfToken      // if
	ElseToken    // else
	ForToken     // for
	InToken      // in
	TrueToken    // true
	FalseToken   // false
	keywordEnd   // End of keyword tokens

	operatorBegin       // Start of operator tokens
	LTToken             // <
	GTToken             // >
	GTEToken            // >=
	LTEToken            // <=
	EQToken             // ==
	NEQToken            // !=
	NotToken            // !
	DotToken            // .
	DotDotToken         // ..
	AndToken            // &&
	OrToken             // ||
	LPARENToken         // (
	RPARENToken         // )
	LBracketToken       // [
	RBracketToken       // ]
	COMMAToken          // ,
	LBRACEToken         // {
	RBRACEToken         // }
	SlashToken          // /
	StartTagOpenToken   // <
	TagCloseToken       // >
	TagSelfClosingToken // />
	EndTagOpenToken     // </
	AttrValSepToken     // =
	SUBToken            // -
	operatorEnd         // End of operator tokens

	TagNameToken      // div
	AttrNameToken     // class
	AttrValDelimToken // ' or "
	AttrValTextToken  // abc

	literalBegin // Start of literal tokens
	IDENTToken   // scoped variable name or component property name
	INTToken     // 123
	FloatToken   // 123.45
	IMAGToken    // 123.4i
	StringToken  // "abc"
	CHARToken    // 'c'
	FMTToken     // YY-MM-DD H:M:S
	literalEnd   // End of literal tokens
)

var tokens = [...]string{
	ErrorToken:   "error",
	EOFToken:     "eof",
	DoctypeToken: "doctype",
	CDATAToken:   "cdata",
	CommentToken: "comment",
	TextToken:    "text",

	StartTagOpenToken:   "<", // <
	TagCloseToken:       ">", // >
	TagSelfClosingToken: "/>",
	EndTagOpenToken:     "</",
	TagNameToken:        "tagName",
	AttrNameToken:       "attributeName",
	AttrValSepToken:     "=",
	AttrValDelimToken:   "attributeValueDelimiter",
	AttrValTextToken:    "attributeValueText",
	SpaceToken:          "space",

	IfToken:    "if",
	ElseToken:  "else",
	ForToken:   "for",
	InToken:    "in",
	TrueToken:  "true",
	FalseToken: "false",

	LTToken:       "<",
	GTToken:       ">",
	GTEToken:      ">=",
	LTEToken:      "<=",
	EQToken:       "==",
	NEQToken:      "!=",
	NotToken:      "!",
	DotToken:      ".",
	DotDotToken:   "..",
	AndToken:      "&&",
	OrToken:       "||",
	LPARENToken:   "(",
	RPARENToken:   ")",
	LBracketToken: "[",
	RBracketToken: "]",
	COMMAToken:    ",",
	LBRACEToken:   "{",
	RBRACEToken:   "}",
	SlashToken:    "/",
	SUBToken:      "-",

	IDENTToken:  "identifier", // macro variable name or component property name
	INTToken:    "integer",    // 123
	FloatToken:  "float",      // 123.45
	StringToken: "string",     // "abc"
	FMTToken:    "formatting",
}

var keywords = map[string]Token{
	"if":    IfToken,
	"else":  ElseToken,
	"for":   ForToken,
	"in":    InToken,
	"true":  TrueToken,
	"false": FalseToken,
}

func (tok Token) String() string {
	s := ""
	if 0 <= int(tok) && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

func (tok Token) IsLiteral() bool { return literalBegin < tok && tok < literalEnd }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (tok Token) IsOperator() bool {
	return operatorBegin < tok && tok < operatorEnd
}

// IsKeyword returns true for tokens corresponding to keywords;
// it returns false otherwise.
func (tok Token) IsKeyword() bool { return keywordBegin < tok && tok < keywordEnd }

// Lookup maps an identifier to its keyword token or [IDENT] (if not a keyword).
func Lookup(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return IDENTToken
}

func IsAsciiLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func IsXmlName(name string) bool {
	if name == "" {
		return false
	}
	length := len(name)
	for i, c := range name {
		if i == 0 && !IsAsciiLetter(c) {
			return false
		}
		if !IsAsciiLetter(c) && c != ':' && c != '-' && !unicode.IsDigit(c) {
			return false
		}
		if i == length && (c == ':' || c == '-') {
			return false
		}
	}
	return true
}

// IsKeyword reports whether name is a Vanilla keyword, such as "if" or "for".
func IsKeyword(name string) bool {
	_, ok := keywords[name]
	return ok
}

// IsIdentifier reports whether name is a Vanilla identifier, that is, a non-empty
// string made up of letters, digits, and underscores, where the first character
// is not a digit. Keywords are not identifiers.
func IsIdentifier(name string) bool {
	if name == "" || IsKeyword(name) {
		return false
	}
	for i, c := range name {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return true
}
