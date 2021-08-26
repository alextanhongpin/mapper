package internal

import (
	"fmt"

	"github.com/alextanhongpin/mapper"
	. "github.com/dave/jennifer/jen"
)

// FuncBuilder builds local functions in a function scope.
type FuncBuilder struct {
	resolver Resolver

	// The parent function.
	fn *mapper.Func
}

func NewFuncBuilder(r Resolver, fn *mapper.Func) *FuncBuilder {
	return &FuncBuilder{
		resolver: r,
		fn:       fn,
	}
}

func (b *FuncBuilder) genReturnOnError(fn *mapper.Func) *Statement {
	// TODO: Turn into error.
	if fn.Error == nil {
		panic(fmt.Sprintf("mapper: missing return error for %s", b.fn.PrettySignature()))
	}
	return If(Id("err").Op("!=").Id("nil")).Block(ReturnFunc(func(g *Group) {
		// Output:
		// if err != nil {
		//   return B{}, err
		// }
		g.Add(List(GenTypeName(fn.To.Type).Values(), Id("err")))
	}))
}

func (b *FuncBuilder) GenReturnOnError() *Statement {
	return b.genReturnOnError(b.fn)
}

func (b *FuncBuilder) BuildFuncCall(c *C, fn *mapper.Func, lhs, rhs *mapper.Type) {
	b.buildFunc(c, fn, lhs, rhs, b.genFuncCall(fn, lhs, rhs))
}

func (b *FuncBuilder) BuildMethodCall(c *C, prefix *Statement, fn *mapper.Func, lhs, rhs *mapper.Type) {
	b.buildFunc(c, fn, lhs, rhs, b.genMethodCall(prefix, fn, lhs, rhs))
}

func (b *FuncBuilder) genMethodCall(prefix *Statement, method *mapper.Func, lhs, rhs *mapper.Type) *Statement {
	var (
		r                    = b.resolver
		a0Selection          = r.RhsVar
		requiresInputPointer = !lhs.IsPointer && method.From.Type.IsPointer
		requiresInputValue   = lhs.IsPointer && !method.From.Type.IsPointer
	)
	return prefix.Clone().Dot(method.Name).Call(Do(func(s *Statement) {
		if requiresInputPointer {
			// Output:
			// fn.Fn(&a0Name)
			s.Add(Op("&"))
		} else if requiresInputValue {
			// Output:
			// fn.Fn(*a0Name)
			s.Add(Op("*"))
		}
	}).Add(a0Selection())).Clone()
}

func (b *FuncBuilder) genFuncCall(fn *mapper.Func, lhs, rhs *mapper.Type) *Statement {
	var (
		r                    = b.resolver
		a0Selection          = r.RhsVar
		requiresInputPointer = fn.RequiresInputPointer(lhs)
		requiresInputValue   = fn.RequiresInputValue(lhs)
	)

	// Output:
	// fn.Fn(a0Name)
	return Qual(fn.PkgPath, fn.Name).Call(Do(func(s *Statement) {
		if requiresInputPointer {
			// Output:
			// fn.Fn(&a0Name)
			s.Add(Op("&"))
		} else if requiresInputValue {
			// Output:
			// fn.Fn(*a0Name)
			s.Add(Op("*"))
		}
	}).Add(a0Selection())).Clone()
}

func (b *FuncBuilder) buildFunc(c *C, fn *mapper.Func, lhs, rhs *mapper.Type, fnCall *Statement) {
	var (
		r           = b.resolver
		a0Name      = r.LhsVar
		a0Selection = r.RhsVar
		to          = fn.To.Type
	)
	defer func() {
		r.Assign()
	}()

	//g.validateFunctionSignatureMatch(fn, lhs, rhs)

	inputIsPointer := lhs.IsPointer
	outputIsPointer := fn.To.Type.IsPointer

	hasError := fn.Error != nil
	if !hasError {
		// NOT SLICE.
		if !lhs.IsSlice {
			// INPUT IS POINTER.
			if inputIsPointer {
				// Output:
				// var a0Name *fn.T
				// if a0.Name != nil {
				// 	 a0Name = fn.Fn(a0.Name)  // Funn requires pointer input
				// 	 a0Name = fn.Fn(*a0.Name) // Func requires value input
				// }
				c.Add(Var().Add(a0Name()).Add(GenType(to)))
				c.Add(If(a0Selection().Op("!=").Id("nil")).Block(
					a0Name().Op("=").Add(fnCall.Clone())),
				)
				return
			}
			// INPUT IS NOT POINTER.
			// Output:
			// a0Name := fn.Fn(&a0.Name) // Func requires pointer input.
			// a0Name := fn.Fn(a0.Name)  // Func requires value input.
			c.Add(a0Name().Op(":=").Add(fnCall.Clone()))
			return
		}

		// IS SLICE + INPUT IS POINTER
		if inputIsPointer {
			// OUTPUT IS POINTER.
			if outputIsPointer {
				// Output:
				// var a0Name []b.B
				// for _, each := range a0.Name {
				//   if each != nil {
				//     tmp := fn.Fn(*each)
				//     if tmp != nil {
				//       a0Name = append(a0Name, *tmp) // Expects value return.
				//       a0Name = append(a0Name, tmp)  // Expects pointer return.
				//     }
				//   }
				// }
				return
			}
			// OUTPUT IS NOT POINTER.
			// Output:
			// var a0Name []b.B
			// for _, each := range a0.Name {
			//   if each != nil {
			//     tmp := fn.Fn(*each)
			//     a0Name = append(a0Name, &tmp) // Expects pointer return.
			//     a0Name = append(a0Name, tmp)  // Expects value return.
			//   }
			// }
			return
		}

		// IS SLICE + INPUT IS NOT POINTER + OUTPUT IS POINTER
		if outputIsPointer {
			// Output:
			// var a0Name []b.B
			// for i, each := range a0.Name {
			//   tmp := fn.Fn(&each)
			//   if tmp != nil {
			//     a0Name = append(a0Name, tmp)  // Expects output pointer.
			//     a0Name = append(a0Name, *tmp) // Expects output value.
			//   }
			// }
		}

		// IS SLICE + INPUT IS NOT POINTER + OUTPUT IS NOT POINTER
		// Output:
		// a0Name := make([]b.B, len(a0.Name))
		// for i, each := range a0.Name {
		//   tmp := fn.Fn(&each)
		//   a0Name[i] = &tmp // Expects output pointer.
		//   a0Name[i] = tmp  // Expects output value.
		// }
		return // END OF NOT SLICE
	} // END OF HAS ERROR

	// HAS ERROR.
	if !lhs.IsSlice {
		if inputIsPointer {
			// Output:
			// var a1Name *fn.T
			// if a0Name != nil {
			//   a1Name, err = fn.Fn(*a0Name)
			//   if err != nil {
			//      return nil, err
			//   }
			// }
			c.Add(Var().Add(a0Name()).Add(GenType(to)))
			c.Add(If(a0Selection().Op("!=").Id("nil")).Block(
				List(a0Name(), Id("err")).Op("=").Add(b.GenReturnOnError()),
				If(Id("err").Op("!=").Id("nil")).Block(),
			))
			return
		}
		// Output:
		// a2Name, err := fn.Fn(a1Name)
		// if err != nil {
		//  return nil, err
		// }
		c.Add(
			List(a0Name(), Id("err")).Op("=").Add(fnCall.Clone()),
			If(Id("err").Op("!=").Id("nil")).Block(b.GenReturnOnError()),
		)
		return
	}
}
