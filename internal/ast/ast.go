package ast

import (
	"github.com/supaleon/vanilla/internal/token"
)

// Node is the interface of all node type.
type Node interface {
	Range() token.Range
}
