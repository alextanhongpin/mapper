package internal

import (
	"fmt"

	"github.com/alextanhongpin/mapper"
	. "github.com/dave/jennifer/jen"
)

func GenReturnTypeOnError(fn mapper.Func) *Statement {
	// TODO: Turn into error.
	if fn.Error == nil {
		panic(fmt.Sprintf("mapper: missing return error for %s", fn.PrettySignature()))
	}
	return If(Id("err").Op("!=").Id("nil")).Block(ReturnFunc(func(g *Group) {
		// Output:
		// if err != nil {
		//   return &B{}, err
		// }
		out := fn.To.Type
		g.Add(List(Do(func(s *Statement) {
			if out.IsStruct {
				if out.IsPointer {
					s.Add(Op("&").Add(GenTypeName(fn.To.Type))).Values()
				} else {
					s.Add(GenTypeName(fn.To.Type)).Values()
				}
			} else if out.IsPointer || out.IsSlice {
				s.Add(Id("nil"))
			}
		}), Id("err")))
	}))
}

func GenReturnTypeNameOnError(fn mapper.Func) *Statement {
	// TODO: Turn into error.
	if fn.Error == nil {
		panic(fmt.Sprintf("mapper: missing return error for %s", fn.PrettySignature()))
	}
	return If(Id("err").Op("!=").Id("nil")).Block(ReturnFunc(func(g *Group) {
		// Output:
		// if err != nil {
		//   return B{}, err
		// }
		g.Add(List(GenTypeName(fn.To.Type).Values(), Id("err")))
	}))
}
