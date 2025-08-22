package ast

import (
	"github.com/supaleon/vanilla/internal/token"
)

type ImportKind int

const (
	// ImportSTMT aka `import xx`
	ImportSTMT ImportKind = iota
	// ImportDynamic aka `import()`
	ImportDynamic
)

type ImportSpec struct {
	Kind ImportKind
}

type ESModule struct {
	Imports []*ImportSpec
}

func (e *ESModule) Range() token.Range {
	return token.Range{}
}
