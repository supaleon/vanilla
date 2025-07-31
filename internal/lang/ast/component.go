package ast

import (
	"github.com/supaleon/vanilla/internal/lang/token"
)

type Template struct {
}

func (t *Template) Range() token.Range {
	//todo
	return token.Range{}
}

type Component struct {
	ESModule *ESModule
	Template *Template
}

func (c *Component) Range() token.Range {
	//todo
	return token.Range{}
}
