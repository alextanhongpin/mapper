package internal

import (
	"fmt"

	"github.com/alextanhongpin/mapper"
	. "github.com/dave/jennifer/jen"
)

func GenInputType(fn *mapper.Func) *Statement {
	return Id(argsWithIndex(fn.From.Name, 0)).Add(GenType(fn.From.Type))
}

func GenReturnType(fn *mapper.Func) *Statement {
	if fn.Error {
		return Parens(List(GenType(fn.To.Type), Id("error")))
	}
	return GenType(fn.To.Type)
}

func GenReturnValue(fn *mapper.Func) *Statement {
	if !fn.Error {
		panic(fmt.Sprintf("mapper: missing return error for %s", fn.Signature()))
	}

	return If(Err().Op("!=").Nil()).Block(ReturnFunc(func(g *Group) {
		// Output:
		//
		// if err != nil {
		//   return &B{}, err
		// }
		out := fn.To.Type
		g.Add(List(Do(func(s *Statement) {
			if out.IsPointer || out.IsSlice {
				s.Add(Nil())
			} else {
				s.Add(GenTypeName(out)).Values()
			}
		}), Err()))
	}))
}
