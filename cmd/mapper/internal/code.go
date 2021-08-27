package internal

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type C []jen.Code

func (c *C) Add(code ...jen.Code) *C {
	*c = append(*c, code...)
	return c
}

func (c *C) String() string {
	res := jen.Null()
	for i, code := range *c {
		res.Add(code)
		if i != len(*c)-1 {
			res.Add(jen.Line())
		}
	}
	return fmt.Sprintf("%#v", res)
}
