package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
	. "github.com/dave/jennifer/jen"
)

// FuncBuilder builds local functions in a function scope.
type FuncBuilder struct {
	resolver Resolver

	// The parent function.
	fn *mapper.Func
}

// NewFuncBuilder returns a pointer for FuncBuilder.
func NewFuncBuilder(r Resolver, fn *mapper.Func) *FuncBuilder {
	return &FuncBuilder{
		resolver: r,
		fn:       fn,
	}
}

func (b *FuncBuilder) GenReturnType() *Statement {
	return GenReturnType(b.fn)
}

func (b *FuncBuilder) GenReturnOnError() *Statement {
	return GenReturnValue(b.fn)
}

func (b *FuncBuilder) BuildFuncCall(fn *mapper.Func, lhs, rhs types.Type) *Statement {
	return b.buildFunc(fn, lhs, rhs, func(assign *Statement, op string) *Statement {
		prefix := Qual(fn.PkgPath, fn.Name)
		return b.genMethodCall(prefix, assign, op, fn, lhs, rhs)
	})
}

func (b *FuncBuilder) BuildMethodCall(prefix *Statement, fn *mapper.Func, lhs, rhs types.Type) *Statement {
	return b.buildFunc(fn, lhs, rhs, func(assign *Statement, op string) *Statement {
		return b.genMethodCall(prefix, assign, op, fn, lhs, rhs)
	})
}

func (b *FuncBuilder) genMethodCall(prefix, assign *Statement, op string, method *mapper.Func, lhs, rhs types.Type) *Statement {
	var (
		r                    = b.resolver
		a0Selection          = r.RhsVar
		requiresInputPointer = method.RequiresInputPointer(lhs)
		requiresInputValue   = method.RequiresInputValue(lhs)
	)
	fnCall := prefix.Clone().Call(Do(func(s *Statement) {
		if requiresInputPointer {
			// Output:
			// fn.Fn(&a0Name)
			s.Add(Op("&"))
		} else if requiresInputValue {
			// Output:
			// fn.Fn(*a0Name)
			s.Add(Op("*"))
		}

		if mapper.IsSlice(lhs) && mapper.IsSlice(rhs) && !mapper.IsSlice(method.From.Type) {
			s.Add(Id("each"))
		} else {
			s.Add(a0Selection())
		}
	}))

	if !method.Error {
		return assign.Clone().Op(op).Add(fnCall)
	}

	return CustomFunc(Options{Multi: true}, func(g *Group) {
		if op == "=" {
			g.Add(Var().Err().Error())
		}
		g.Add(List(assign.Clone(), Err()).Op(op).Add(fnCall))
		g.Add(b.GenReturnOnError())
	})
}

func (b *FuncBuilder) buildFunc(fn *mapper.Func, lhs, rhs types.Type, fnAssignment func(*Statement, string) *Statement) *Statement {
	var (
		r      = b.resolver
		a0Name = r.LhsVar
	)
	defer func() {
		r.Assign()
	}()

	ins := mapper.IsSlice(lhs)
	args := mapper.IsSlice(fn.From.Type)
	outs := mapper.IsSlice(rhs)

	// struct2struct
	if ins == outs && ins == args {
		return struct2Struct(r, fn, lhs, rhs, fnAssignment)
	} else if ins == outs && ins != args {
		if !ins {
			panic("unhandled method call")
		}
		return slice2slice(r, fn, lhs, rhs, fnAssignment)
	} else {
		/*
			Output:

			a0Name := fn.Fn(a0.Name)
		*/
		return fnAssignment(a0Name(), ":=")
	}
}

func struct2Struct(r Resolver, fn *mapper.Func, lhs, rhs types.Type, fnAssignment func(*Statement, string) *Statement) *Statement {
	var (
		a0Name      = r.LhsVar
		a0Selection = r.RhsVar
	)

	inp := mapper.IsPointer(lhs)
	outp := mapper.IsPointer(rhs)
	argp := mapper.IsPointer(fn.From.Type)
	resp := mapper.IsPointer(fn.To.Type)

	var nilHandler func(*Statement) *Statement
	if inp != argp && inp {
		nilHandler = func(s *Statement) *Statement {
			return Custom(Options{Multi: true},
				/*
					Output:

					var a0Name (*)b.B
					if a0.Name != nil {
						// To be filled ...
					}
				*/
				Var().Add(a0Name()).Add(GenType(rhs)),
				If(a0Selection().Op("!=").Nil()).Block(s),
			)
		}
	}

	if outp != resp && outp {
		if nilHandler != nil {
			return nilHandler(Custom(Options{Multi: true},
				/*
					Output:

					var a0Name (*)b.B
					if a0.Name != nil {
						tmp := fnpkg.Fn(a0.Name)
						a0Name = &tmp
					}
				*/
				fnAssignment(Id("tmp"), ":="),
				a0Name()).Op("=").Op("&").Id("tmp"),
			)
		}

		return CustomFunc(Options{Multi: true}, func(g *Group) {
			/*
				Output:

				a0Name := fnpkg.Fn(a0.Name)
				a1Name := &a0Name
			*/
			g.Add(fnAssignment(a0Name(), ":="))
			r.Assign()
			g.Add(a0Name().Op(":=").Op("&").Add(a0Selection()))
		})
	}

	if nilHandler != nil {
		/*
			Output:

			var a0Name (*)b.B
			if a0.Name != nil {
				a0Name = fnpkg.Fn(a0.Name)
			}
		*/
		return nilHandler(fnAssignment(a0Name(), "="))
	}

	/*
		Output:

		a0Name := fnpkg.Fn(a0.Name)
	*/
	return fnAssignment(a0Name(), ":=")
}

func slice2slice(r Resolver, fn *mapper.Func, lhs, rhs types.Type, fnAssignment func(*Statement, string) *Statement) *Statement {
	var (
		a0Name      = r.LhsVar
		a0Selection = r.RhsVar
	)

	inp := mapper.IsPointer(lhs)
	argp := mapper.IsPointer(fn.From.Type)
	outp := mapper.IsPointer(rhs)
	resp := mapper.IsPointer(fn.To.Type)

	if inp != argp && inp {
		if outp != resp && resp {
			/*
				Output:

				var a0Name []b.B
				for _, each := range a0.Name {
					if each != nil {
						tmp := fn.Fn(*each)
						if tmp != nil {
							a0Name = append(a0Name, *tmp) // Expects value return.
							a0Name = append(a0Name, tmp) 	// Expects pointer return.
						}
					}
				}
			*/
			return Custom(Options{Multi: true},
				Var().Add(a0Name()).Add(GenType(rhs)),
				For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
					If(Id("each").Op("!=").Nil().Block(
						fnAssignment(Id("tmp"), "="),
						If(Id("tmp").Op("!=").Nil()).Block(
							a0Selection().Op("=").Append(a0Selection(), Do(func(s *Statement) {
								if outp != resp {
									if outp {
										s.Add(Op("&"))
									} else {
										s.Add(Op("*"))
									}
								}
							}).Id("tmp")),
						),
					)),
				),
			)
		}

		/*
			Output:

			var a0Name []b.B
			for _, each := range a0.Name {
				if each != nil {
					tmp := fn.Fn(*each)
					a0Name = append(a0Name, *tmp) // Expects value return.
					a0Name = append(a0Name, tmp) 	// Expects pointer return.
				}
			}
		*/
		return Custom(Options{Multi: true},
			Var().Add(a0Name()).Add(GenType(rhs)),
			For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
				If(Id("each").Op("!=").Nil().Block(
					fnAssignment(Id("tmp"), "="),
					a0Selection().Op("=").Append(a0Selection(), Do(func(s *Statement) {
						if outp == resp {
							return
						}

						if outp {
							s.Add(Op("&"))
						} else {
							s.Add(Op("*"))
						}
					}).Id("tmp")),
				)),
			),
		)
	}

	if resp {
		/*
			Condition: in is not pointer, out is pointer.
			Output:

			var a0Name [](*)b.B
			for _, each := range a0.Name {
				tmp := pkgfn.Fn(each)
				if tmp != nil {
					a0Name = append(a0Name, &tmp) // Expects output pointer for value result.
					a0Name = append(a0Name, tmp)  // Expects output pointer for pointer result.
					a0Name = append(a0Name, *tmp) // Expects output value for pointer result.
				}
			}
		*/
		return Custom(Options{Multi: true},
			Var().Add(a0Name()).Add(GenType(rhs)),
			For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
				fnAssignment(Id("tmp"), "="),
				If(Id("tmp").Op("!=").Nil()).Block(
					a0Name().Op("=").Append(a0Name(), Do(func(s *Statement) {
						if resp == outp {
							return
						}

						if outp {
							s.Add(Op("&"))
						} else {
							s.Add(Op("*"))
						}
					}).Id("tmp")),
				),
			),
		)
	}

	/*
		Condition: in/out is not pointer.
		Output:

		a0Name := make([]b.B, a0.Name)
		for i, each := range a0.Name {
			a0Name[i] = pkgfn.Fn(each)
		}
	*/
	return Custom(Options{Multi: true},
		a0Name().Op(":=").Make(Add(GenType(rhs)), Len(a0Selection())),
		For(List(Id("i"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
			fnAssignment(a0Name().Index(Id("i")), "="),
		),
	)
}
