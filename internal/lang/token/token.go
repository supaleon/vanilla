package token

import (
	"strconv"
	"unicode"
)

// Token is the set of lexical tokens of the Vanilla component programming language.
type Token uint8

const (
	IllegalToken  Token = iota // Special token: syntax error.
	EOFToken                   // Special token: end of a file.
	CommentToken               // <!--x-->
	DoctypeToken               // <!DOCTYPE x>
	PainTextToken              // abc

	EndTagToken         // </
	SelfClosingTagToken // />

	keywordBegin // Start of keyword tokens
	IfToken      // if
	ElseToken    // else
	ForToken     // for
	InToken      // in
	TrueToken    // true
	FalseToken   // false
	DeferToken   // defer
	ContextToken // context
	keywordEnd   // End of keyword tokens

	operatorBegin // Start of operator tokens
	LTToken       // <
	GTToken       // >
	GTEToken      // >=
	LTEToken      // <=
	EQToken       // ==
	NEQToken      // !=
	NotToken      // !
	DotToken      // .
	DotDotToken   // ..
	AndToken      // &&
	OrToken       // ||
	LPARENToken   // (
	RPARENToken   // )
	LBracketToken // [
	RBracketToken // ]
	COMMAToken    // ,
	LBRACEToken   // {
	RBRACEToken   // }
	ColonToken    // :
	REMToken      // %
	SlashToken    // /
	operatorEnd   // End of operator tokens

	literalBegin // Start of literal tokens
	IDENTToken   // div,MyComponent
	INTToken     // 123
	FloatToken   // 123.45
	NumberToken  // 123
	StringToken  // "abc"
	literalEnd   // End of literal tokens
)

var tokens = [...]string{
	IllegalToken:  "illegal",
	EOFToken:      "eof",
	DoctypeToken:  "doctype",
	CommentToken:  "comment",
	PainTextToken: "painText",

	EndTagToken:         "</",
	SelfClosingTagToken: "/>",

	IfToken:      "if",
	ElseToken:    "else",
	DeferToken:   "defer",
	ForToken:     "for",
	ContextToken: "context",
	InToken:      "in",
	TrueToken:    "true",
	FalseToken:   "false",

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
	ColonToken:    ":",
	SlashToken:    "/",
	REMToken:      "%",

	IDENTToken:  "identifier", // div, MyComponent
	INTToken:    "integer",    // 123
	FloatToken:  "float",      // 123.45
	NumberToken: "number",
	StringToken: "string", // "abc"
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token, keywordEnd-(keywordBegin+1))
	for i := keywordBegin + 1; i < keywordEnd; i++ {
		keywords[tokens[i]] = i
	}
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

// Lookup returns the keyword or identifier with the given name.
func Lookup(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return IDENTToken
}

func IsKeyword(name string) bool {
	_, ok := keywords[name]
	return ok
}

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
