package internal

import (
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

func (b *FuncBuilder) BuildFuncCall(c *C, fn *mapper.Func, lhs, rhs *mapper.Type) {
	b.buildFunc(c, fn, lhs, rhs, func() *Statement {
		prefix := Qual(fn.PkgPath, fn.Name)
		return b.genMethodCall(prefix, fn, lhs, rhs)
	})
}

func (b *FuncBuilder) BuildMethodCall(c *C, prefix *Statement, fn *mapper.Func, lhs, rhs *mapper.Type) {
	b.buildFunc(c, fn, lhs, rhs, func() *Statement {
		return b.genMethodCall(prefix, fn, lhs, rhs)
	})
}

func (b *FuncBuilder) genMethodCall(prefix *Statement, method *mapper.Func, lhs, rhs *mapper.Type) *Statement {
	var (
		r                    = b.resolver
		a0Selection          = r.RhsVar
		requiresInputPointer = method.RequiresInputPointer(lhs)
		requiresInputValue   = method.RequiresInputValue(lhs)
	)
	return prefix.Clone().Call(Do(func(s *Statement) {
		if requiresInputPointer {
			// Output:
			// fn.Fn(&a0Name)
			s.Add(Op("&"))
		} else if requiresInputValue {
			// Output:
			// fn.Fn(*a0Name)
			s.Add(Op("*"))
		}

		if lhs.IsSlice && rhs.IsSlice {
			s.Add(Id("each"))
		} else {
			s.Add(a0Selection())
		}
	}))
}

func (b *FuncBuilder) buildFunc(c *C, fn *mapper.Func, lhs, rhs *mapper.Type, fnCall func() *Statement) {
	var (
		r           = b.resolver
		a0Name      = r.LhsVar
		a0Selection = r.RhsVar
	)
	defer func() {
		r.Assign()
	}()

	inputIsPointer := lhs.IsPointer
	outputIsPointer := fn.To.Type.IsPointer
	expectsPointer := rhs.IsPointer
	inputIsSlice := lhs.IsSlice
	outputIsSlice := rhs.IsSlice

	if !fn.Error {
		if !inputIsSlice {
			if inputIsPointer {
				if !outputIsPointer && expectsPointer {
					// NO ERROR > NOT SLICE > INPUT IS POINTER > OUTPUT IS NOT POINTER > EXPECTS POINTER
					//
					// Output:
					// var a0Name *fn.T
					// if a0.Name != nil {
					//   tmp := fn.Fn(a0.Name)  // Func requires pointer input
					//   a0Name = &tmp
					// }
					c.Add(
						Var().Add(a0Name()).Add(GenType(rhs)),
						If(a0Selection().Op("!=").Nil()).Block(
							Id("tmp").Op(":=").Add(fnCall()),
							a0Name().Op("=").Op("&").Id("tmp"),
						),
					)
				} else {
					// NO ERROR > NOT SLICE > INPUT IS POINTER > OUTPUT IS NOT POINTER > DOES NOT EXPECT POINTER
					// NO ERROR > NOT SLICE > INPUT IS POINTER > OUTPUT IS POINTER > EXPECT POINTER
					// NO ERROR > NOT SLICE > INPUT IS POINTER > OUTPUT IS POINTER > DOES NOT EXPECT POINTER (WILL PANIC)
					//
					// Output:
					// var a0Name fn.T
					// if a0.Name != nil {
					//   a0Name = fn.Fn(a0.Name)  // Func requires pointer input
					// 	 a0Name = fn.Fn(*a0.Name) // Func requires value input
					// }
					c.Add(
						Var().Add(a0Name()).Add(GenType(rhs)),
						If(a0Selection().Op("!=").Nil()).Block(
							a0Name().Op("=").Add(fnCall()),
						),
					)
				}
			} else {
				if !outputIsPointer && expectsPointer {
					// NO ERROR > NOT SLICE > INPUT IS NOT POINTER > OUTPUT IS POINTER > DOES NOT EXPECT POINTER
					//
					// Output:
					// a0Name := fn.Fn(&a0.Name)
					// a1Name := &a0Name
					c.Add(a0Name().Op(":=").Add(fnCall()))
					r.Assign()
					c.Add(a0Name().Op(":=").Op("&").Add(a0Selection()))
				} else {
					// NO ERROR > NOT SLICE > INPUT IS NOT POINTER > OUTPUT IS POINTER > EXPECTS POINTER
					// NO ERROR > NOT SLICE > INPUT IS NOT POINTER > OUTPUT IS NOT POINTER > DOES NOT EXPECTS POINTER
					// NO ERROR > NOT SLICE > INPUT IS NOT POINTER > OUTPUT IS POINTER > DOES NOT EXPECT POINTER (WILL PANIC ABOVE)
					//
					// Output:
					// a0Name := fn.Fn(&a0.Name) // Func requires pointer input.
					// a0Name := fn.Fn(a0.Name)  // Func requires value input.
					c.Add(a0Name().Op(":=").Add(fnCall()))
				}
			}
		} else {
			if !outputIsSlice {
				// INPUT IS SLICE > OUTPUT IS NOT SLICE > INPUT IS NOT POINTER > OUTPUT IS NOT POINTER
				//
				/*
					Output:

					a0Name := fn.Fn(a0.Name)
				*/
				c.Add(
					a0Name().Op(":=").Add(fnCall()),
				)
			} else {
				// NO ERROR > IS SLICE
				if inputIsPointer {
					// IS SLICE > INPUT IS POINTER
					//
					// Output:
					// var a0Name []*b.B // If output is pointer
					// var a0Name []b.B  // If output is value
					// for _, each := range a0.Name {
					//   if each != nil {
					//     tmp := fn.Fn(*each)
					//     if tmp != nil {
					//       a0Name = append(a0Name, *tmp) // Expects value return.
					//       a0Name = append(a0Name, tmp)  // Expects pointer return.
					//     }
					//   }
					// }
					c.Add(
						Var().Add(a0Name()).Add(GenType(rhs)),
						For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
							If(Id("each").Op("!=").Nil().Block(
								Id("tmp").Op(":=").Add(fnCall()),
								If(Id("tmp").Op("!=").Nil()).Block(
									a0Selection().Op("=").Append(a0Selection(), Do(func(s *Statement) {
										if outputIsPointer && !expectsPointer {
											s.Add(Op("*"))
										}
										if !outputIsPointer && expectsPointer {
											s.Add(Op("&"))
										}
									}).Id("tmp")),
								),
							)),
						),
					)
				} else {
					if outputIsPointer {
						// IS SLICE > INPUT IS NOT POINTER > OUTPUT IS POINTER
						//
						// Output:
						// var a0Name []b.B  // Expects value.
						// var a0Name []*b.B // Expects pointer.
						// for _, each := range a0.Name {
						//   tmp := fn.Fn(&each)
						//   if tmp != nil {
						//     a0Name = append(a0Name, &tmp) // Expects output pointer for value result.
						//     a0Name = append(a0Name, tmp)  // Expects output pointer for pointer result.
						//     a0Name = append(a0Name, *tmp) // Expects output value for pointer result.
						//   }
						// }
						c.Add(
							Var().Add(a0Name()).Add(GenType(rhs)),
							For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
								Id("tmp").Op(":=").Add(fnCall()),
								If(Id("tmp").Op("!=").Nil()).Block(
									a0Name().Op("=").Append(a0Name(), Do(func(s *Statement) {
										if outputIsPointer && !expectsPointer {
											s.Add(Op("*"))
										} else if !outputIsPointer && expectsPointer {
											s.Add(Op("&"))
										}
									}).Id("tmp")),
								),
							),
						)
					} else {
						// IS SLICE > INPUT IS NOT POINTER > OUTPUT IS NOT POINTER
						//
						// Output:
						// a0Name := make([]b.B, a0.Name)
						// for i, each := range a0.Name {
						//   a0Name[i] = fn.Fn(&each)
						// }
						c.Add(
							a0Name().Op(":=").Make(Add(GenType(rhs)), Len(a0Selection())),
							For(List(Id("i"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
								a0Name().Index(Id("i")).Op("=").Add(fnCall()),
							),
						)
					}
				}
			}
		}

	} else {
		// HAS ERROR.
		if !inputIsSlice {
			// IS NOT SLICE
			if inputIsPointer {
				if !outputIsPointer && expectsPointer {
					// HAS ERROR > IS NOT SLICE > INPUT IS POINTER > OUTPUT IS NOT POINTER > EXPECTS POINTER
					//
					// Output:
					// var a1Name *fn.T
					// if a0Name != nil {
					//   tmp, err := fn.Fn(&a0Name)
					//   if err != nil {
					//      return nil, err
					//   }
					//   a1Name = &tmp
					// }
					c.Add(
						Var().Add(a0Name()).Add(GenType(rhs)),
						If(a0Selection().Op("!=").Nil()).Block(
							List(Id("tmp"), Err()).Op(":=").Add(fnCall()),
							b.GenReturnOnError(),
							a0Name().Op("=").Op("&").Id("tmp"),
						),
					)
				} else {
					// HAS ERROR > IS NOT SLICE > INPUT IS POINTER > OUTPUT IS NOT POINTER > DOES NOT EXPECT POINTER
					// HAS ERROR > IS NOT SLICE > INPUT IS POINTER > OUTPUT IS NOT POINTER > EXPECTS POINTER
					// HAS ERROR > IS NOT SLICE > INPUT IS POINTER > OUTPUT IS POINTER > DOES NOT EXPECT POINTER (WILL PANIC)
					//
					// Output:
					// var a1Name *fn.T
					// if a0Name != nil {
					//   a1Name, err = fn.Fn(&a0Name)
					//   if err != nil {
					//      return nil, err
					//   }
					// }
					c.Add(Var().Add(a0Name()).Add(GenType(rhs)))
					c.Add(If(a0Selection().Op("!=").Nil()).Block(
						List(a0Name(), Err()).Op("=").Add(fnCall()),
						b.GenReturnOnError(),
					))
				}
			} else {
				if !outputIsPointer && expectsPointer {
					// HAS ERROR > IS NOT SLICE > INPUT IS NOT POINTER  > OUTPUT IS NOT POINTER > EXPȨCTS POINTER
					// Output:
					//
					// a0Name, err := fn.Fn(a1Name)
					// if err != nil {
					//   return nil, err
					// }
					// a1Name := &a0Name
					c.Add(
						List(a0Name(), Err()).Op(":=").Add(fnCall()),
						b.GenReturnOnError(),
					)
					r.Assign()
					c.Add(a0Name().Op(":=").Op("&").Add(a0Selection()))
				} else {
					// HAS ERROR > IS NOT SLICE > INPUT IS NOT POINTER
					// Output:
					//
					// a2Name, err := fn.Fn(a1Name)
					// if err != nil {
					//   return nil, err
					// }
					c.Add(
						List(a0Name(), Err()).Op(":=").Add(fnCall()),
						b.GenReturnOnError(),
					)
				}
			}
		} else {
			// IS SLICE
			if inputIsPointer {
				if outputIsPointer {
					// HAS ERROR > IS SLICE > INPUT IS POINTER > OUTPUT IS POINTER
					//
					// Output:
					// var a0Name []*b.B // If output is pointer
					// var a0Name []b.B  // If output is value
					// for _, each := range a0.Name {
					//   if each != nil {
					//     tmp, err := fn.Fn(*each)
					//     if err != nil {
					//       return nil, err
					//     }
					//     if tmp != nil {
					//       a0Name = append(a0Name, *tmp) // Expects value return.
					//       a0Name = append(a0Name, tmp)  // Expects pointer return.
					//     }
					//   }
					// }
					c.Add(
						Var().Add(a0Name()).Add(GenType(rhs)),
						For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
							If(Id("each").Op("!=").Nil().Block(
								List(Id("tmp"), Err()).Op(":=").Add(fnCall()),
								b.GenReturnOnError(),
								If(Id("tmp").Op("!=").Nil()).Block(
									a0Selection().Op("=").Append(a0Selection(), Do(func(s *Statement) {
										if outputIsPointer && !expectsPointer {
											s.Add(Op("*"))
										}
										if !outputIsPointer && expectsPointer {
											s.Add(Op("&"))
										}
									}).Id("tmp")),
								),
							)),
						),
					)
				} else {
					// HAS ERROR > IS SLICE > INPUT IS POINTER > OUTPUT IS POINTER
					//
					// Output:
					// var a0Name []*b.B // If output is pointer
					// var a0Name []b.B  // If output is value
					// for _, each := range a0.Name {
					//   if each != nil {
					//     tmp, err := fn.Fn(*each)
					//     if err != nil {
					//       return nil, err
					//     }
					//     a0Name = append(a0Name, *tmp) // Expects value return.
					//     a0Name = append(a0Name, tmp)  // Expects pointer return.
					//   }
					// }
					c.Add(
						Var().Add(a0Name()).Add(GenType(rhs)),
						For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
							If(Id("each").Op("!=").Nil().Block(
								List(Id("tmp"), Err()).Op(":=").Add(fnCall()),
								b.GenReturnOnError(),
								If(Id("tmp").Op("!=").Nil()).Block(
									a0Selection().Op("=").Append(a0Selection(), Do(func(s *Statement) {
										if outputIsPointer && !expectsPointer {
											s.Add(Op("*"))
										}
										if !outputIsPointer && expectsPointer {
											s.Add(Op("&"))
										}
									}).Id("tmp")),
								),
							)),
						),
					)
				}
			} else {
				if outputIsPointer {
					// HAS ERROR > IS SLICE > INPUT IS NOT POINTER > OUTPUT IS POINTER
					//
					// Output:
					// var a0Name []b.B  // Expects value.
					// var a0Name []*b.B // Expects pointer.
					// for _, each := range a0.Name {
					//   tmp, err := fn.Fn(&each)
					//   if err != nil {
					//     return nil, err
					//   }
					//   if tmp != nil {
					//     a0Name = append(a0Name, &tmp) // Expects output pointer for value result.
					//     a0Name = append(a0Name, tmp)  // Expects output pointer for pointer result.
					//     a0Name = append(a0Name, *tmp) // Expects output value for pointer result.
					//   }
					// }
					c.Add(
						Var().Add(a0Name()).Add(GenType(rhs)),
						For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
							List(Id("tmp"), Err()).Op(":=").Add(fnCall()),
							b.GenReturnOnError(),
							If(Id("tmp").Op("!=").Nil()).Block(
								a0Name().Op("=").Append(a0Name(), Do(func(s *Statement) {
									if outputIsPointer && !expectsPointer {
										s.Add(Op("*"))
									} else if !outputIsPointer && expectsPointer {
										s.Add(Op("&"))
									}
								}).Id("tmp")),
							),
						),
					)
				} else {
					// HAS ERROR > IS SLICE > INPUT IS NOT POINTER > OUTPUT IS NOT POINTER
					//
					// Output:
					// var a0Name []b.B  // Expects value.
					// var a0Name []*b.B // Expects pointer.
					// for _, each := range a0.Name {
					//   tmp, err := fn.Fn(&each)
					//   if err != nil {
					//     return nil, err
					//   }
					//   a0Name = append(a0Name, &tmp) // Expects output pointer for value result.
					//   a0Name = append(a0Name, tmp)  // Expects output pointer for pointer result.
					//   a0Name = append(a0Name, *tmp) // Expects output value for pointer result.
					// }
					c.Add(
						Var().Add(a0Name()).Add(GenType(rhs)),
						For(List(Id("_"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
							List(Id("tmp"), Err()).Op(":=").Add(fnCall()),
							b.GenReturnOnError(),
							a0Name().Op("=").Append(a0Name(), Do(func(s *Statement) {
								if outputIsPointer && !expectsPointer {
									s.Add(Op("*"))
								} else if !outputIsPointer && expectsPointer {
									s.Add(Op("&"))
								}
							}).Id("tmp")),
						),
					)
				}
			}
		}
	}
}
