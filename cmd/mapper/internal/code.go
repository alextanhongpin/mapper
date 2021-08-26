package internal

import "github.com/dave/jennifer/jen"

type C []jen.Code

func (c *C) Add(code ...jen.Code) *C {
	*c = append(*c, code...)
	return c
}
