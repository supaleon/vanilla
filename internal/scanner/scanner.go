// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package scanner converts a source file to a stream of tokens.
// Unlike a textbook-style lexer, the scanner implementation violates many programming guidelines. For example,
// the scanner switches states not only in the main loop but also within the scanXx series of functions,
// reducing the overhead of repeated checks. The scanner also heavily duplicates code instead of reusing it.
// While this increases maintenance difficulty, it is justified by the performance gains.

package scanner

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/supaleon/vanilla/internal/token"
)

type state uint8

const (
	stateText      state = iota // abc
	stateCodeBlock              // {...}

	stateTagOpen  // < or </
	stateStartTag // div

	stateAttrName          // class
	stateAttrValSep        // =
	stateAttrExpr          // {!disable}
	stateUnquotedAttrVal   // data-value=123
	stateQuotedAttrVal     // "dark {isLoggedIn:pro}"
	stateAttrValDelimOpen  // ' or "
	stateAttrValInterp     // {isLoggedIn:dark}
	stateAttrValDelimClose // ' or "

	stateEndTag       // div
	stateTagClose     // >
	stateTagSelfClose // />
)

// class={!article.archived} -> scanAttrExpr
// class="{!article.archived}", class='{!article.archived}', -> scanAttrValue -> attrValText, attValInterp
// {article.archived} -> scanTextExpr

type ErrorHandler func(pos token.Position, msg string)

type Scanner struct {
	// immutable state
	file *token.File // source file handle
	src  []byte      // source

	ch       rune // current character
	offset   int  // character offset
	rdOffset int  // reading offset (position after current character)
	lbOffset int  // current line break offset

	rawTag []byte // current tag: title,textarea,style,script,plaintext,xmp...

	state            state
	attrValDelimOpen rune // attribute value attrValDelimOpen ' or "

	errorHandler ErrorHandler // error reporting; or nil
	// public state - ok to modify
	errorCount int // number of errors encountered

	debug bool
}

const (
	bom = 0xFEFF // byte order mark, only permitted as very first character
	eof = -1     // end of file
)

func (s *Scanner) scanRawText(tag []byte) (lit string) {
	off := s.offset
	l := len(tag)
	for {
		s.next()
		if s.ch < 0 {
			break
		}
		if s.ch == '<' {
			p := s.peek()
			if p == '/' {
				buf, size := s.peekN(l + 1)
				if size > 0 && bytes.Equal(buf[1:], tag) {
					break
				}
			}
		}
	}
	lit = string(s.src[off:s.offset])
	s.rawTag = nil
	return
}

// Read the next Unicode char into s.ch.
// s.ch < 0 means end-of-file.
//
// For optimization, there is some overlap between this method and
// s.scanIdentifier.
func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lbOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lbOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = eof
	}
}

// peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, peek returns 0.
func (s *Scanner) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

func (s *Scanner) peekN(n int) (data []byte, size int) {
	l := s.rdOffset + n
	if l < len(s.src) {
		return s.src[s.rdOffset:l], n
	}
	return
}

func (s *Scanner) peekRune() (char rune, size int) {
	if s.rdOffset < len(s.src) {
		char, size = utf8.DecodeRune(s.src[s.rdOffset:])
		return
	}
	return eof, 0
}

func New(file *token.File, src []byte, err ErrorHandler) (s *Scanner) {
	// Explicitly initialize all fields since a scanner may be reused.
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s = &Scanner{}
	s.file = file
	s.src = src
	s.errorHandler = err

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lbOffset = 0
	s.errorCount = 0

	s.next()
	if s.ch == bom {
		s.next() // ignore BOM at file beginning
	}
	return
}

func (s *Scanner) error(offs int, msg string) {
	if s.errorHandler != nil {
		s.errorHandler(s.file.Position(s.file.Location(offs)), msg)
	}
	s.errorCount++
}

func (s *Scanner) errorf(offs int, format string, args ...any) {
	s.error(offs, fmt.Sprintf(format, args...))
}

//var prefix = []byte("line ")

// updateLineInfo parses the incoming comment text at offset offs
// as a line directive. If successful, it updates the line info table
// for the position next per the line directive.
func (s *Scanner) updateLineInfo(next, offs int, text []byte) {
	// extract comment text
	if text[1] == '*' {
		text = text[:len(text)-2] // lop off trailing "*/"
	}
	text = text[7:] // lop off leading "//line " or "/*line "
	offs += 7

	i, n, ok := trailingDigits(text)
	if i == 0 {
		return // ignore (not a line directive)
	}
	// i > 0

	if !ok {
		// text has a suffix :xxx but xxx is not a number
		s.error(offs+i, "invalid line number: "+string(text[i:]))
		return
	}

	// Put a cap on the maximum size of line and column numbers.
	// 30 bits allows for some additional space before wrapping an int32.
	// Keep this consistent with cmd/compile/internal/syntax.PosMax.
	const maxLineCol = 1 << 30
	var line, col int
	i2, n2, ok2 := trailingDigits(text[:i-1])
	if ok2 {
		//line filename:line:col
		i, i2 = i2, i
		line, col = n2, n
		if col == 0 || col > maxLineCol {
			s.error(offs+i2, "invalid column number: "+string(text[i2:]))
			return
		}
		text = text[:i2-1] // lop off ":col"
	} else {
		//line filename:line
		line = n
	}

	if line == 0 || line > maxLineCol {
		s.error(offs+i, "invalid line number: "+string(text[i:]))
		return
	}

	// If we have a column (//line filename:line:col form),
	// an empty filename means to use the previous filename.
	filename := string(text[:i-1]) // lop off ":line", and trim white space
	if filename == "" && ok2 {
		filename = s.file.Position(s.file.Location(offs)).Filename
	} else if filename != "" {
		// Put a relative filename in the current directory.
		// This is for compatibility with earlier releases.
		// See issue 26671.
		//filename = filepath.Clean(filename)
		//if !filepath.IsAbs(filename) {
		//	filename = filepath.Join(s.dir, filename)
		//}
	}

	s.file.AddLineColumnInfo(next, filename, line, col)
}

func trailingDigits(text []byte) (int, int, bool) {
	i := bytes.LastIndexByte(text, ':') // look from right (Windows filenames may contain ':')
	if i < 0 {
		return 0, 0, false // no ":"
	}
	// i >= 0
	n, err := strconv.ParseUint(string(text[i+1:]), 10, 0)
	return i + 1, int(n), err == nil
}

func litname(prefix rune) string {
	switch prefix {
	case 'x':
		return "hexadecimal literal"
	case 'o', '0':
		return "octal literal"
	case 'b':
		return "binary literal"
	}
	return "decimal literal"
}

// digits accepts the sequence { digit | '_' }.
// If base <= 10, digits accepts any decimal digit but records
// the offset (relative to the source start) of a digit >= base
// in *invalid, if *invalid < 0.
// digits returns a bitset describing whether the sequence contained
// digits (bit 0 is set), or separators '_' (bit 1 is set).
func (s *Scanner) digits(base int, invalid *int) (digsep int) {
	if base <= 10 {
		max := rune('0' + base)
		for isDecimal(s.ch) || s.ch == '_' {
			ds := 1
			if s.ch == '_' {
				ds = 2
			} else if s.ch >= max && *invalid < 0 {
				*invalid = s.offset // record invalid rune offset
			}
			digsep |= ds
			s.next()
		}
	} else {
		for isHex(s.ch) || s.ch == '_' {
			ds := 1
			if s.ch == '_' {
				ds = 2
			}
			digsep |= ds
			s.next()
		}
	}
	return
}

// invalidSep returns the index of the first invalid separator in x, or -1.
func invalidSep(x string) int {
	x1 := ' ' // prefix char, we only care if it's 'x'
	d := '.'  // digit, one of '_', '0' (a digit), or '.' (anything else)
	i := 0

	// a prefix counts as a digit
	if len(x) >= 2 && x[0] == '0' {
		x1 = lower(rune(x[1]))
		if x1 == 'x' || x1 == 'o' || x1 == 'b' {
			d = '0'
			i = 2
		}
	}

	// mantissa and exponent
	for ; i < len(x); i++ {
		p := d // previous digit
		d = rune(x[i])
		switch {
		case d == '_':
			if p != '0' {
				return i
			}
		case isDecimal(d) || x1 == 'x' && isHex(d):
			d = '0'
		default:
			if p == '_' {
				return i - 1
			}
			d = '.'
		}
	}
	if d == '_' {
		return len(x) - 1
	}

	return -1
}

func (s *Scanner) scanNumber() (tok token.Token, lit string) {
	offs := s.offset
	tok = token.ILLEGAL

	base := 10        // number base
	prefix := rune(0) // one of 0 (decimal), '0' (0-octal), 'x', 'o', or 'b'
	digsep := 0       // bit 0: digit present, bit 1: '_' present
	invalid := -1     // index of invalid digit in literal, or < 0

	// integer part
	if s.ch != '.' {
		tok = token.INT
		if s.ch == '0' {
			s.next()
			switch lower(s.ch) {
			case 'x':
				s.next()
				base, prefix = 16, 'x'
			case 'o':
				s.next()
				base, prefix = 8, 'o'
			case 'b':
				s.next()
				base, prefix = 2, 'b'
			default:
				base, prefix = 8, '0'
				digsep = 1 // leading 0
			}
		}
		digsep |= s.digits(base, &invalid)
	}

	// fractional part
	if s.ch == '.' {
		if p := s.peek(); p == '.' {
			tok = token.INT
			lit = string(s.src[offs:s.offset])
			return
		}
		tok = token.FLOAT
		if prefix == 'o' || prefix == 'b' {
			s.error(s.offset, "invalid radix point in "+litname(prefix))
		}
		s.next()

		digsep |= s.digits(base, &invalid)
	}

	if digsep&1 == 0 {
		s.error(s.offset, litname(prefix)+" has no digits")
	}

	// exponent
	if e := lower(s.ch); e == 'e' || e == 'p' {
		switch {
		case e == 'e' && prefix != 0 && prefix != '0':
			s.errorf(s.offset, "%q exponent requires decimal mantissa", s.ch)
		case e == 'p' && prefix != 'x':
			s.errorf(s.offset, "%q exponent requires hexadecimal mantissa", s.ch)
		}
		s.next()
		tok = token.FLOAT
		if s.ch == '+' || s.ch == '-' {
			s.next()
		}
		ds := s.digits(10, nil)
		digsep |= ds
		if ds&1 == 0 {
			s.error(s.offset, "exponent has no digits")
		}
	} else if prefix == 'x' && tok == token.FLOAT {
		s.error(s.offset, "hexadecimal mantissa requires a 'p' exponent")
	}

	// suffix 'i'
	if s.ch == 'i' {
		tok = token.ILLEGAL
		s.error(s.offset, "imaginary numbers are not allowed")
		s.next()
	}

	lit = string(s.src[offs:s.offset])
	if tok == token.INT && invalid >= 0 {
		s.errorf(invalid, "invalid digit %q in %s", lit[invalid-offs], litname(prefix))
	}
	if digsep&2 != 0 {
		if i := invalidSep(lit); i >= 0 {
			s.error(offs+i, "'_' must separate successive digits")
		}
	}

	return tok, lit
}

// scanIdentifier reads the string of valid identifier characters at s.offset.
// It must only be called when s.ch is known to be a valid letter.
//
// Be careful when making changes to this function: it is optimized and affects
// scanning performance significantly.
func (s *Scanner) scanIdentifier() string {
	offs := s.offset

	// Optimize for the common case of an ASCII identifier.
	//
	// Ranging over s.src[s.rdOffset:] lets us avoid some bounds checks, and
	// avoids conversions to runes.
	//
	// In case we encounter a non-ASCII character, fall back on the slower path
	// of calling into s.next().
	for rdOffset, b := range s.src[s.rdOffset:] {
		if 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || b == '_' || '0' <= b && b <= '9' {
			// Avoid assigning a rune for the common case of an ascii character.
			continue
		}
		s.rdOffset += rdOffset
		if 0 < b && b < utf8.RuneSelf {
			// Optimization: we've encountered an ASCII character that's not a letter
			// or number. Avoid the call into s.next() and corresponding set up.
			//
			// Note that s.next() does some line accounting if s.ch is '\n', so this
			// shortcut is only possible because we know that the preceding character
			// is not '\n'.
			s.ch = rune(b)
			s.offset = s.rdOffset
			s.rdOffset++
			goto exit
		}
		// We know that the preceding character is valid for an identifier because
		// scanIdentifier is only called when s.ch is a letter, so calling s.next()
		// at s.rdOffset resets the scanner state.
		s.next()
		for isUnicodeLetter(s.ch) || isDigit(s.ch) {
			s.next()
		}
		goto exit
	}
	s.offset = len(s.src)
	s.rdOffset = len(s.src)
	s.ch = eof

exit:
	return string(s.src[offs:s.offset])
}

// scanEscape parses an escape sequence where rune is the accepted
// escaped attrValDelimOpen. In case of a syntax error, it stops at the offending
// character (without consuming it) and returns false. Otherwise,
// it returns true.
func (s *Scanner) scanEscape(quote rune) bool {
	offs := s.offset

	var n int
	var base, max uint32
	switch s.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		s.next()
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		s.next()
		n, base, max = 2, 16, 255
	case 'u':
		s.next()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		s.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		msg := "unknown escape sequence"
		if s.ch < 0 {
			msg = "escape sequence not terminated"
		}
		s.error(offs, msg)
		return false
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(s.ch))
		if d >= base {
			msg := fmt.Sprintf("illegal character %#U in escape sequence", s.ch)
			if s.ch < 0 {
				msg = "escape sequence not terminated"
			}
			s.error(s.offset, msg)
			return false
		}
		x = x*base + d
		s.next()
		n--
	}

	if x > max || 0xD800 <= x && x < 0xE000 {
		s.error(offs, "escape sequence is invalid Unicode code point")
		return false
	}

	return true
}

func (s *Scanner) scanRune() (tok token.Token, lit string) {
	// '\'' opening already consumed
	offs := s.offset - 1

	valid := true
	n := 0
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			// only report error if we don't have one already
			if valid {
				s.error(offs, "rune literal not terminated")
				valid = false
			}
			break
		}
		s.next()
		if ch == '\'' {
			break
		}
		n++
		if ch == '\\' {
			if !s.scanEscape('\'') {
				valid = false
			}
			// continue to read to closing quote
		}
	}

	if valid && n != 1 {
		s.error(offs, "illegal rune literal")
	}
	tok = token.CHAR
	lit = string(s.src[offs:s.offset])
	return
}

// scanString scan characters in if statement.
// {if var1 == "abc"}
// Some case like this {if var1 == "<a"} conflicts with the HTML start tag open token.
// Fixme: Use {if var1 == "&lt;abc"} instead to avoid ambiguity???
//
// Some case like this {if var1 < 42} conflicts with the VSCode Editor's builtin HTML parser, but ok.
// Use {if 42 > var1} instead to avoid ambiguity.
func (s *Scanner) scanString() (tok token.Token, lit string) {
	// '"' opening already consumed
	off := s.offset - 1
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(off, "string literal not terminated")
			break
		}
		if ch == '"' {
			s.next()
			lit = string(s.src[off:s.offset])
			tok = token.STRING
			break
		}
		if ch == '\\' {
			s.scanEscape('"')
		}
		s.next()
	}
	return
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) scanXmlInstruction() (tok token.Token, lit string) {
	tok = token.COMMENT
	off := s.offset
	// consume '<'
	s.next()
	for {
		s.next()
		if s.ch < 0 {
			break
		}
		if s.ch == '>' {
			s.next()
			break
		}
	}
	lit = string(s.src[off:s.offset])
	s.error(off, "component source code cannot contain XML processing instructions")
	return
}

var doctype = []byte("DOCTYPE")

// scanComment treats all characters starting with `<!` as comments,
// except those starting with `<!DOCTYPE`
func (s *Scanner) scanComment() (tok token.Token, lit string) {
	tok = token.COMMENT
	off := s.offset
	err := "incorrectly opened comment"
	// consume '<'
	s.next()

	switch r, _ := s.peekRune(); {
	case r == '-':
		// consume '!'
		s.next()
		if p := s.peek(); p == '-' {
			s.next()
			s.next()
			tok, lit = s.readComment(off)
			return
		}
	case r == '[':
		// <![CDATA[section]]>
		err = "component source code cannot contain XML CDATA"
	case lower(r) == 'd':
		// TODO use a hash algo do this.
		if dt, size := s.peekN(7); size > 0 && bytes.EqualFold(dt, doctype) {
			tok = token.DOCTYPE
			// <!DOCTYPE html>
			err = "component source code cannot contain HTML Doctype"
		}
	}

	for {
		s.next()
		if s.ch < 0 {
			break
		}
		if s.ch == '>' {
			s.next()
			break
		}
	}
	lit = string(s.src[off:s.offset])
	s.error(off, err)
	return
}

// readComment
// <!--abc--> comment.
func (s *Scanner) readComment(off int) (tok token.Token, lit string) {
	tok = token.COMMENT
	// consume characters until '-->' or eof found
	for {
		s.next()
		ch := s.ch
		if ch < 0 {
			break
		}
		// '-->'
		if ch == '-' {
			s.next()
			if s.ch == '-' {
				if p := s.peek(); p == '>' {
					s.next()
					s.next()
					break
				}
			}
		}
	}
	lit = string(s.src[off:s.offset])
	return
}

func (s *Scanner) scanText() (tok token.Token, lit string) {
	tok = token.TEXT
	off := s.offset
	// scan until found <, {, eof
	for {
		// always ignore the first char.
		s.next()
		ch := s.ch
		// escape.
		if s.ch == '\\' {
			if p := s.peek(); p == '{' || p == '}' {
				// consume '\\'
				s.next()
				// consume '{' or '}'
				s.next()
				continue
			}
		}
		if ch < 0 || ch == '{' {
			lit = string(s.src[off:s.offset])
			break
		}
		// treat unmatched '}' as regular text but report it as an error since it's missing its opening '{'
		if s.ch == '}' {
			s.error(s.offset, "code block closing character '}' is missing opening character '{'")
			s.next()
			continue
		}
		if s.advanceToTagOpen(true) {
			lit = string(s.src[off:s.offset])
			break
		}
	}
	return
}

func (s *Scanner) advanceToTagOpen(skipWhitespace bool) (ok bool) {
	if skipWhitespace {
		s.skipWhitespace()
	}
	// <div or </div
	if s.ch == '<' {
		s.state = stateTagOpen
	}
	//todo
	s.attrValDelimOpen = 0
	return
}

func (s *Scanner) advanceToTagClose(skipWhitespace bool) (ok bool) {
	if skipWhitespace {
		s.skipWhitespace()
	}
	switch ch := s.ch; {
	case ch < 0:
		ok = true
	case ch == '>':
		s.state = stateTagClose
		ok = true
	}
	return
}

// advance to code sync point.
func (s *Scanner) advance(skipWhitespace bool) (ok bool) {
	if skipWhitespace {
		s.skipWhitespace()
	}
	switch ch := s.ch; {
	case ch < 0:
		ok = true
	case ch == '>':
		s.state = stateTagClose
		ok = true
	case ch == '<':
		s.state = stateTagOpen
		ok = true
	case ch == '/':
		if p := s.peek(); p == '>' {
			s.state = stateTagSelfClose
			ok = true
		}
	}
	return
}

func (s *Scanner) scanStartTag() (lit string) {
	// '<' already consumed
	off := s.offset
	errOff := -1
	for {
		// the first char is always a valid Unicode letter.
		s.next()
		// <div>
		// <div class="abc">
		if s.ch < 0 || s.ch == '>' || isWhitespace(s.ch) {
			break
		}
		// <br/>
		if s.ch == '/' {
			if p := s.peek(); p == '>' {
				break
			}
		}
		// TODO: strict mode?
		if !isTagNameChar(s.ch) {
			if errOff < 0 {
				errOff = s.offset
			}
		}
	}

	buf := s.src[off:s.offset]
	lit = string(buf)
	if IsRawTag(lit) {
		s.rawTag = buf
	}
	// reports error.
	if errOff > -1 {
		r, _ := utf8.DecodeRune(s.src[errOff:])
		s.errorf(errOff, "invalid character %q in start tag name", r)
	} else if IsDeprecatedTag(lit) {
		s.errorf(off, "%q is deprecated", lit)
	}
	return
}

// scanEndTag accepts characters starting with any letters except whitespaces and '>'.
func (s *Scanner) scanEndTag() (lit string) {
	// `</` already consumed.
	off := s.offset
	errOff := -1
	for {
		// the first char is always a valid Unicode letter.
		s.next()
		if s.ch < 0 {
			lit = string(s.src[off:s.offset])
			break
		}
		// TODO: strict mode?
		if isTagNameChar(s.ch) {
			continue
		}
		// maybe `</div x >` or `</div  >`
		if isWhitespace(s.ch) {
			lit = string(s.src[off:s.offset])
			// consume all the rest characters.
			for {
				s.next()
				if s.ch == '>' {
					s.state = stateTagClose
					break
				}
			}
			break
		}

		if s.ch == '>' {
			lit = string(s.src[off:s.offset])
			s.state = stateTagClose
			break
		}

		// maybe `</div~abc >` or `</div~ x  >`
		errOff = s.offset
	}

	// reports error.
	if errOff > -1 {
		r, _ := utf8.DecodeRune(s.src[errOff:])
		s.errorf(errOff, "invalid character %q in end tag name", r)
	}
	return
}

// ending points: whitespace, =, >, />
func (s *Scanner) scanAttrName() string {
	off := s.offset
	errOff := -1
	for {
		if s.ch < 0 || s.ch == '>' || isWhitespace(s.ch) {
			break
		}
		// <br hash/>
		if s.ch == '/' {
			if r, _ := s.peekRune(); r == '>' {
				break
			}
		}
		if s.ch == '=' {
			// maybe <div =abc >
			if s.offset != off {
				s.state = stateAttrValSep
				break
			}
		}

		if !isAttrNameChar(s.ch) {
			if errOff < 0 {
				errOff = s.offset
			}
		}
		s.next()
	}
	// reports the error.
	if errOff > -1 {
		r, _ := utf8.DecodeRune(s.src[errOff:])
		s.errorf(errOff, "invalid character %q in attribute name", r)
	}

	return string(s.src[off:s.offset])
}

func (s *Scanner) scanUnquotedAttrValue() (tok token.Token, lit string) {
	errOff := -1
	tok = token.ATTRValText
	off := s.offset
	for {
		// class=}
		// class="
		// class='
		// class==
		if s.ch == '"' || s.ch == '\'' || s.ch == '=' || s.ch == '}' {
			if errOff < 0 {
				errOff = s.offset
			}
			s.next()
			continue
		}
		if s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
			s.state = stateAttrName
			break
		}
		if s.advance(false) {
			break
		}
		s.next()
	}

	// reports the error.
	if errOff > -1 {
		r, _ := utf8.DecodeRune(s.src[errOff:])
		s.errorf(errOff, "invalid character %q in unquoted attribute value", r)
	}
	lit = string(s.src[off:s.offset])
	return
}

func (s *Scanner) switchAttrValState() {
	// NB: maybe <div class=  "cls">, syntactically this is allowed
	s.skipWhitespace()
	// `=` already consumed.
	switch ch := s.ch; {
	case ch == '"' || ch == '\'':
		s.state = stateAttrValDelimOpen
	case ch == '{':
		s.state = stateAttrExpr
	case ch == '>': // <div class= >
		s.state = stateTagClose
		s.error(s.offset, "missing attribute value")
	case ch == '/':
		if p := s.peek(); p == '>' {
			s.state = stateTagClose
			s.error(s.offset, "missing attribute value")
		}
		// fixme.
		fallthrough
	default:
		s.state = stateUnquotedAttrVal
	}
}

func (s *Scanner) scanQuotedAttrVal() (tok token.Token, lit string) {
	tok = token.ATTRValText
	off := s.offset
	// maybe "{var1} abc {var2}"
	for {
		// the first char cannot be {, }, '
		s.next()
		// println("========", string(s.ch))
		if s.ch < 0 {
			//println("========")
			tok = token.ILLEGAL
			s.error(s.offset, "attribute not terminated")
			break
		}
		// `'` or `"`
		if s.ch == s.attrValDelimOpen {
			s.state = stateAttrValDelimClose
			break
		}

		if s.ch == '\\' {
			s.next()
			if s.ch == '{' || s.ch == '}' {
				s.next()
				continue
			}
		}

		if s.ch == '{' || s.ch == '}' {
			s.state = stateAttrValInterp
			break
		}
	}
	lit = string(s.src[off:s.offset])
	return
}

// scanBasicExpr never returns token.ILLEGAL.
func (s *Scanner) scanBasicExpr() (tok token.Token, lit string) {
	tok = token.ILLEGAL
	switch ch := s.ch; {
	case ch == '-':
		tok = token.SUB
		s.next()
	//case ch == '+':
	//	tok = token.ADD
	//	s.next()
	case isDecimal(ch) || ch == '.' && isDecimal(rune(s.peek())):
		tok, lit = s.scanNumber()
	case isUnicodeLetter(ch):
		lit = s.scanIdentifier()
		if len(lit) > 1 {
			// keywords are longer than one letter - avoid lookup otherwise
			tok = token.Lookup(lit)
		} else {
			tok = token.IDENT
		}
	case ch == '!':
		s.next()
		tok = token.NOT
	case ch == '.':
		tok = token.DOT
		s.next()
	case ch == '[':
		tok = token.LBRACKET
		s.next()
	case ch == ']':
		tok = token.RBRACKET
		s.next()
	case ch == '(':
		tok = token.LPAREN
		s.next()
	case ch == ')':
		tok = token.RPAREN
		s.next()
	}
	return
}

// scanSpecifier
// conditional text expression & format expression.
// class="{!disable:light}"
// <div>{isLoggedIn: Welcome back!}</div>
// class="{user.createTime % YY-MM-DD HH:MM:SS}"
// class="{user.score %.2f}"
func (s *Scanner) scanSpecifier(specTok token.Token) (tok token.Token, lit string) {
	tok = token.ILLEGAL
	off := s.offset
	errOff := -1
	for {
		s.next()
		if s.ch < 0 {
			errOff = s.offset
			break
		}
		if s.ch == '}' {
			tok = specTok
			break
		}
		// "
		if s.state == stateQuotedAttrVal && s.ch == s.attrValDelimOpen {
			s.state = stateAttrValDelimClose
			errOff = s.offset
			break
		}
		// <
		if s.state == stateText && s.ch == '<' {
			s.state = stateTagOpen
			errOff = s.offset
			break
		}
	}
	lit = string(s.src[off:s.offset])
	if errOff > -1 {
		s.error(errOff, "conditional text expression not terminated")
	}
	return
}

// scanAttrExpr parses attribute expressions or property assignment expressions.
// <input type="checkbox" checked={user.isDeactivated}/>
// <button disabled={!user.isLoggedIn}>Submit</button>
// <Component tags={user.Tags} />
func (s *Scanner) scanAttrExpr() (tok token.Token, lit string) {
	tok = token.ILLEGAL
	off := s.offset
	switch ch := s.ch; {
	case isWhitespace(ch):
		s.error(s.offset, "whitespace is not allowed in attribute expression")
		lit = string(s.ch)
		s.next()
		s.state = stateAttrName
	case ch == '{':
		// fixme: class={{{{hello}
		tok = token.LBRACE
		s.next()
	case ch == '}':
		s.next()
		tok = token.RBRACE
		s.state = stateAttrName
		if s.ch > -1 && s.ch != '>' && s.ch != '/' && !isWhitespace(s.ch) {
			s.error(s.offset, "missing whitespace between attribute name and the previous attribute expression")
		}
	default:
		tok, lit = s.scanBasicExpr()
		// continue scan...
		if tok == token.ILLEGAL {
			r, _ := utf8.DecodeRune(s.src[s.offset:])
			s.errorf(s.offset, "invalid character %q in attribute expression", r)
			// scan until find a sync point: whitespace, >,
			for {
				s.next()
				if s.ch < 0 || s.ch == '>' {
					break
				}
				if isWhitespace(s.ch) {
					s.state = stateAttrName
					break
				}
				// />
				if s.ch == '/' {
					s.next()
					if s.ch == '>' {
						break
					}
				}
				if s.ch == '}' {
					break
				}
			}
			lit = string(s.src[off:s.offset])
		}
	}
	return
}

// scanAttrValInterp parses attribute value interpolation expressions.
func (s *Scanner) scanAttrValInterp() (tok token.Token, lit string) {
	off := s.offset
	switch ch := s.ch; {
	case ch == '{':
		s.next()
		tok = token.LBRACE
	case ch == '}':
		s.next()
		tok = token.RBRACE
		s.state = stateQuotedAttrVal
	case ch == '%':
		tok, lit = s.scanSpecifier(token.FMT)
	case ch == ':':
		tok, lit = s.scanSpecifier(token.CONDText)
	default:
		tok, lit = s.scanBasicExpr()
		if tok == token.ILLEGAL {
			r, _ := utf8.DecodeRune(s.src[s.offset:])
			s.errorf(s.offset, "invalid character %q in attribute interpolation expression", r)
			// scan until find a sync point: }, ', "
			for {
				s.next()
				if s.ch == s.attrValDelimOpen {
					s.state = stateAttrValDelimClose
					break
				}
				if s.ch < 0 || s.ch == '}' {
					break
				}
			}
			lit = string(s.src[off:s.offset])
		}
	}
	return
}

func (s *Scanner) scanCodeBlock() (tok token.Token, lit string) {
	// {!disable}
	tok = token.ILLEGAL
	off := s.offset
	switch ch := s.ch; {
	case ch == '{':
		s.next()
		tok = token.LBRACE
	case ch == '}':
		s.next()
		tok = token.RBRACE
		s.state = stateText
	case ch == '/':
		tok = token.SLASH
		s.next()
		if isWhitespace(s.ch) {
			for {
				s.next()
				if s.ch < 0 || s.ch == '}' || s.advanceToTagOpen(true) {
					break
				}
			}
			lit = string(s.src[off:s.offset])
			tok = token.ILLEGAL
			s.error(off, "invalid flow control end token")
		}
	case ch == '-':
		tok = token.SUB
		s.next()
	case isDecimal(ch) || ch == '.' && isDecimal(rune(s.peek())):
		tok, lit = s.scanNumber()
	case isUnicodeLetter(ch):
		lit = s.scanIdentifier()
		if len(lit) > 1 {
			// keywords are longer than one letter - avoid lookup otherwise
			tok = token.Lookup(lit)
		} else {
			tok = token.IDENT
		}
	case ch == '.':
		tok = token.DOT
		s.next()
		if s.ch == '.' {
			s.next()
			tok = token.DOTDot
		}
	case ch == '=':
		s.next()
		if s.ch == '=' {
			s.next()
			tok = token.EQ
			break
		}
		// fixme: caution!
		tok = token.ILLEGAL
		for {
			s.next()
			if s.ch < 0 || s.advanceToTagOpen(false) {
				lit = string(s.src[off:s.offset])
				break
			}
		}
	case ch == '>':
		if s.ch == '=' {
			s.next()
			tok = token.GE
		} else {
			tok = token.GT
		}
	case ch == '<':
		s.next()
		if s.ch == '=' {
			s.next()
			tok = token.LE
		} else {
			//println("11111========", string(s.ch))
			tok = token.LT
		}
	case ch == '%':
		tok, lit = s.scanSpecifier(token.FMT)
	case ch == ':':
		tok, lit = s.scanSpecifier(token.CONDText)
	case ch == '!':
		s.next()
		tok = token.NOT
	case ch == '[':
		tok = token.LBRACKET
		s.next()
	case ch == ']':
		tok = token.RBRACKET
		s.next()
	case ch == '(':
		tok = token.LPAREN
		s.next()
	case ch == ')':
		tok = token.RPAREN
		s.next()
	case ch == ',':
		s.next()
		tok = token.COMMA
	case ch == '"':
		s.next()
		tok, lit = s.scanString()
	case ch == '\'':
		s.next()
		tok, lit = s.scanRune()
	default:
		// handler error token until found } or html TagOpen sync points.
		for {
			s.next()
			if s.ch < 0 || s.ch == '}' || s.advanceToTagOpen(false) {
				break
			}
		}
		lit = string(s.src[off:s.offset])
	}
	//println("11111========", string(s.ch), tok.IsKeyword())
	if tok.IsOperator() && !isWhitespace(s.ch) {
		s.error(s.offset, "operator must be surrounded by space")
	}
	return
}

func (s *Scanner) Scan() (loc token.Loc, tok token.Token, lit string) {
	tok = token.ILLEGAL
	if s.offset == 0 && !s.debug {
		// Enforces component source code must begin with a valid HTML tag to ensure readability.
		if s.ch == '<' {
			if r, _ := s.peekRune(); isUnicodeLetter(r) {
				s.state = stateTagOpen
				goto scanAgain
			}
		}
		s.error(s.offset, "component source code must begin with a valid HTML tag")
	}

	if s.state != stateQuotedAttrVal && s.state != stateAttrExpr {
		s.skipWhitespace()
	}

scanAgain:
	if s.ch == eof {
		tok = token.EOF
		return
	}

	loc = s.file.Location(s.offset)

	switch stat := s.state; {
	case stat == stateTagClose:
		// `>`
		s.next()
		tok = token.TAGClose
		s.state = stateText
	case stat == stateTagSelfClose:
		// consume `/>`
		s.next()
		s.next()
		tok = token.TAGSelfClose
		s.state = stateText
	case stat == stateTagOpen:
		r, _ := s.peekRune()
		// end tag open, something like </div
		if r == '/' {
			// consume `<`
			s.next()
			// consume `/`
			s.next()
			s.state = stateEndTag
			tok = token.ENDTagOpen
			// maybe `</🤔`.
			if !isUnicodeLetter(s.ch) {
				s.state = stateText
				s.errorf(s.offset, "invalid character %q in end tag name", s.ch)
			}
			break
		}
		// start tag open, something lik <div
		if isUnicodeLetter(r) {
			// consume `<`
			s.next()
			tok = token.STARTTagOpen
			s.state = stateStartTag
			break
		}
		// maybe `<🤔`, treat as normal text.
		// todo: reports an error?
		tok, lit = s.scanText()
	case stat == stateStartTag:
		tok = token.TAGName
		lit = s.scanStartTag()
		s.state = stateAttrName
	case stat == stateEndTag:
		tok = token.TAGName
		lit = s.scanEndTag()
	case stat == stateAttrName: // from scanStartTag(), scanAttrName(), scanAttrValue, scanQuotedAttrValue
		if s.advance(false) {
			goto scanAgain
		}
		tok = token.ATTRName
		lit = s.scanAttrName()
	case stat == stateAttrValSep:
		// consume '='
		s.next()
		tok = token.ATTRValSep
		s.switchAttrValState()
	case stat == stateAttrExpr:
		tok, lit = s.scanAttrExpr()
	case stat == stateUnquotedAttrVal:
		tok, lit = s.scanUnquotedAttrValue()
	case stat == stateAttrValDelimOpen:
		tok = token.ATTRValDelim
		lit = string(s.ch)
		s.attrValDelimOpen = s.ch
		s.next()
		// class=""
		if s.ch == s.attrValDelimOpen {
			s.state = stateAttrValDelimClose
			break
		}
		s.state = stateQuotedAttrVal
	case stat == stateQuotedAttrVal:
		if s.ch == '{' || s.ch == '}' {
			s.state = stateAttrValInterp
			goto scanAgain
		}
		if s.ch == s.attrValDelimOpen {
			s.state = stateAttrValDelimClose
			goto scanAgain
		}
		tok, lit = s.scanQuotedAttrVal()
	case stat == stateAttrValInterp:
		if s.ch == s.attrValDelimOpen {
			s.state = stateAttrValDelimClose
			goto scanAgain
		}
		tok, lit = s.scanAttrValInterp()
	case stat == stateAttrValDelimClose:
		tok = token.ATTRValDelim
		lit = string(s.ch)
		s.attrValDelimOpen = 0
		s.state = stateAttrName
		s.next()
		if s.advance(false) {
			break
		}
		// class="abc""abc
		if !isWhitespace(s.ch) {
			r, _ := utf8.DecodeRune(s.src[s.offset:])
			s.errorf(s.offset, "missing whitespace between attribute name %q and the previous attribute", r)
		}
	case stat == stateCodeBlock:
		tok, lit = s.scanCodeBlock()
	default:
		if s.rawTag != nil {
			tok = token.TEXT
			lit = s.scanRawText(s.rawTag)
			s.state = stateText
			break
		}
		switch ch := s.ch; {
		case ch == eof:
			tok = token.EOF
		case ch == '<':
			p, _ := s.peekRune()
			// doctype or comment: `<!`
			if p == '!' {
				tok, lit = s.scanComment()
				break
			}
			// xml instruction: `<?`
			if p == '?' {
				tok, lit = s.scanXmlInstruction()
				break
			}
			s.state = stateTagOpen
			goto scanAgain
		case ch == '{' || ch == '}':
			s.state = stateCodeBlock
			goto scanAgain
		default:
			// as text
			tok, lit = s.scanText()
		}
	}
	return
}
