package token

import (
	"strconv"
	"unicode"
)

// Token is the set of lexical tokens of the Vanilla component programming language.
type Token uint8

const (
	ILLEGAL Token = iota // Special token: syntax error.
	EOF                  // Special token: end of a file.
	COMMENT              // <!--x-->
	DOCTYPE              // <!DOCTYPE x>
	CDATA                // <![CDATA[section]]>
	TEXT                 // abc
	SPACE                // ' '

	keywordBegin // Start of keyword tokens
	IF           // if
	ELSE         // else
	FOR          // for
	IN           // in
	TRUE         // true
	FALSE        // false
	NIL          // nil, Go keywords
	LEN          // len()
	OK           // ok()
	EMPTY        // empty()
	keywordEnd   // End of keyword tokens

	operatorBegin // Start of operator tokens
	SUB           // -
	ADD           // +
	LT            // <
	GT            // >
	GE            // >=
	LE            // <=
	EQ            // ==
	NE            // !=
	NOT           // !
	DOT           // .
	DOTDot        // ..
	AND           // &&
	OR            // ||
	LPAREN        // (
	RPAREN        // )
	LBRACKET      // [
	RBRACKET      // ]
	COMMA         // ,
	LBRACE        // {
	RBRACE        // }
	SLASH         // /
	STARTTagOpen  // <
	TAGClose      // >
	TAGSelfClose  // />
	ENDTagOpen    // </
	ATTRValSep    // =
	operatorEnd   // End of operator tokens

	TAGName      // div
	ATTRName     // class
	ATTRValDelim // ' or "
	ATTRValText  // abc

	literalBegin // Start of literal tokens
	IDENT        // scoped variable name or component property name
	INT          // 123
	FLOAT        // 123.45
	STRING       // "abc\n" or `abc\n`
	CHAR         // 'c'
	FMT          // %YY-MM-DD H:M:S or %.3f
	CONDText     // :dark
	literalEnd   // End of literal tokens
)

var tokens = [...]string{
	ILLEGAL: "error",
	EOF:     "eof",
	DOCTYPE: "doctype",
	CDATA:   "cdata",
	COMMENT: "comment",
	TEXT:    "text",

	STARTTagOpen: "<", // <
	TAGClose:     ">", // >
	TAGSelfClose: "/>",
	ENDTagOpen:   "</",
	TAGName:      "tagName",
	ATTRName:     "attributeName",
	ATTRValSep:   "=",
	ATTRValDelim: "attributeValueDelimiter",
	ATTRValText:  "attributeValueText",
	SPACE:        "space",

	IF:    "if",
	ELSE:  "else",
	FOR:   "for",
	IN:    "in",
	TRUE:  "true",
	FALSE: "false",
	NIL:   "nil",   // nil, Go keywords
	LEN:   "len",   // len()
	OK:    "ok",    // ok()
	EMPTY: "empty", // empty()

	LT:       "<",
	GT:       ">",
	GE:       ">=",
	LE:       "<=",
	EQ:       "==",
	NE:       "!=",
	NOT:      "!",
	DOT:      ".",
	DOTDot:   "..",
	AND:      "&&",
	OR:       "||",
	LPAREN:   "(",
	RPAREN:   ")",
	LBRACKET: "[",
	RBRACKET: "]",
	COMMA:    ",",
	LBRACE:   "{",
	RBRACE:   "}",
	SLASH:    "/",
	SUB:      "-",
	ADD:      "+",

	IDENT:    "identifier", // macro variable name or component property name
	INT:      "integer",    // 123
	FLOAT:    "float",      // 123.45
	STRING:   "string",     // "abc"
	CHAR:     "character",
	FMT:      "format",
	CONDText: "condText",
}

var keywords = map[string]Token{
	"if":    IF,
	"else":  ELSE,
	"for":   FOR,
	"in":    IN,
	"true":  TRUE,
	"false": FALSE,
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
	return IDENT
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
