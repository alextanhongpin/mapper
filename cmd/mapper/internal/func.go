package internal

import (
	"fmt"

	"github.com/alextanhongpin/mapper"
	. "github.com/dave/jennifer/jen"
)

func GenInputValue(fn *mapper.Func) *Statement {
	return Id(argsWithIndex(fn.From.Name, 0))
}

func GenInputType(arg *Statement, fn *mapper.Func) *Statement {
	// Output:
	//
	// (a0 ...a.A)
	return arg.Add(GenerateInputType(fn.From.Type, fn.From.Variadic))
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
		g.Add(GenerateOutputType(fn.To.Type, fn.Error))
	}))
}
