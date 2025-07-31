// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copy from Go SDK: go/scanner/scanner.go

package tokenizer

import (
	"fmt"
	"github.com/supaleon/vanilla/internal/lang/token"
	"path/filepath"
	"unicode/utf8"
)

// TODO Error tolerance, support error recovery, continue to scanning

type ErrorHandler func(pos token.Position, msg string)

type state uint8

const (
	inText state = iota
	inTag
)

type Tokenizer struct {
	file *token.File
	dir  string // directory portion of file.Name()
	src  []byte // source
	//source io.Reader // TODO source reader.
	err ErrorHandler // error reporting; or nil

	// scanning state
	ch         rune         // current character
	offset     int          // character offset
	rdOffset   int          // reading offset (position after current character)
	lineOffset int          // current line offset
	nlPos      token.Offset // position of newline in preceding comment

	state state
	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

const (
	bom = 0xFEFF // byte order mark, only permitted as very first character
	eof = -1     // end of file
)

// Read the next Unicode char into s.ch.
// s.ch < 0 means end-of-file.
//
// For optimization, there is some overlap between this method and
// s.scanIdentifier.
func (s *Tokenizer) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
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
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = eof
	}
}

func (s *Tokenizer) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(s.file.Pos(offs)), msg)
	}
	s.ErrorCount++
}

func (s *Tokenizer) errorf(offs int, format string, args ...any) {
	s.error(offs, fmt.Sprintf(format, args...))
}

func (s *Tokenizer) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func NewScanner(file *token.File, src []byte, err ErrorHandler) *Tokenizer {
	s := &Tokenizer{}
	// Explicitly initialize all fields since a scanner may be reused.
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.err = err

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.ErrorCount = 0

	s.next()
	if s.ch == bom {
		s.next() // ignore BOM at file beginning
	}
	return s
}

func (s *Tokenizer) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

func (s *Tokenizer) scanAttrValue() {

}

func (s *Tokenizer) scanAttrKey() {
	//offset := s.offset
	//for {
	//
	//}
}

func (s *Tokenizer) Scan() (pos token.Offset, tok token.Token, lit string) {
	if s.state == inTag {
		s.skipWhitespace()
		s.scanAttrKey()
		// <TAG a=b >
		// 处理 attr
		// 处理 自闭合标签
	}

	for {
		if s.err != nil {
			//println(s.err)
			return
		}
		s.skipWhitespace()
		if s.ch != '<' {
			continue
		}
		// current token start
		pos = s.file.Pos(s.offset)

		//c := s.peek()
		//var tokenType TokenType
		//switch {
		//case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
		//	tokenType = StartTagToken
		//case c == '/':
		//	tokenType = EndTagToken
		//case c == '!' || c == '?':
		//	// We use CommentToken to mean any of "<!--actual comments-->",
		//	// "<!DOCTYPE declarations>" and "<?xml processing instructions?>".
		//	tokenType = CommentToken
		//default:
		//	// Reconsume the current character.
		//	z.raw.end--
		//	continue
		//}
	}
}
