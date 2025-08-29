package scanner

import (
	"unicode"
	"unicode/utf8"
)

// IsRawTag reports whether a tag whose contents are not parsed as HTML.
func IsRawTag(tag string) bool {
	if _, ok := rawTagMap[tag]; ok {
		return true
	}
	return false
}

func IsDeprecatedTag(tag string) bool {
	if _, ok := deprecatedTagMap[tag]; ok {
		return true
	}
	return false
}

func IsVoidTag(tag string) bool {
	if _, ok := voidElementMap[tag]; ok {
		return true
	}
	return false
}

var rawTagMap = map[string]bool{
	"script":    true,
	"style":     true,
	"textarea":  true,
	"title":     true,
	"xmp":       true,
	"plaintext": true,
	"iframe":    true, // srcdoc 属性时
	"noembed":   true,
	"noframes":  true,
	"noscript":  true, // 在支持script的环境中
}

var deprecatedTagMap = map[string]bool{
	"acronym":   true,
	"applet":    true,
	"basefont":  true,
	"bgsound":   true,
	"big":       true,
	"blink":     true,
	"center":    true,
	"dir":       true,
	"font":      true,
	"frame":     true,
	"frameset":  true,
	"isindex":   true,
	"listing":   true,
	"marquee":   true,
	"noframes":  true,
	"plaintext": true,
	"strike":    true,
	"tt":        true,
	"xmp":       true,
}

var voidElementMap = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

func isVisible(ch rune) bool {
	// NB: without space.
	return ch >= 0x21 && ch <= 0x7E
}

func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isUnicodeLetter(ch rune) bool {
	return isLetter(ch) || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isTagNameOpenChar(ch rune) bool {
	// for tag name.
	return isUnicodeLetter(ch)
}

func isTagNameChar(ch rune) bool {
	// for tag name.
	return isUnicodeLetter(ch) || isDigit(ch) || ch == ':' || ch == '-' || ch == '_'
}

func isAttrNameChar(ch rune) bool {
	// for attribute name.
	if ch == '"' || ch == '\'' || ch == '=' || ch == '\\' || ch == '/' {
		return false
	}
	return true
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}

func lower(ch rune) rune { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter

func isHex(ch rune) bool { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }
