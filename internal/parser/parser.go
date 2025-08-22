package parser

type state uint8

const (
	inText state = iota
	inTag
)

type Parser struct {
	//tokenizer *scanner.Scanner
}

//
