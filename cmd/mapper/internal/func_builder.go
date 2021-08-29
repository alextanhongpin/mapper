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

func (b *FuncBuilder) GenReturnType() *Statement {
	if b.fn.Error {
		return Parens(List(GenType(b.fn.To.Type), Id("error")))
	}
	return GenType(b.fn.To.Type)
}

func (b *FuncBuilder) GenReturnOnError() *Statement {
	return GenReturnTypeOnError(*b.fn)
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

		if lhs.IsSlice {
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

	b.validateFunctionSignatureMatch(fn, lhs, rhs)

	inputIsPointer := lhs.IsPointer
	outputIsPointer := fn.To.Type.IsPointer
	expectsPointer := rhs.IsPointer
	if outputIsPointer && !expectsPointer {
		panic("mapper: pointer to value conversion")
	}
	if !fn.Error {
		// NO ERROR
		if !lhs.IsSlice {
			// NOT SLICE
			if inputIsPointer {
				// INPUT IS POINTER
				if outputIsPointer {
					// OUTPUT IS POINTER
					if !expectsPointer {
						// NO ERROR > NOT SLICE > INPUT IS POINTER > OUTPUT IS POINTER > DOES NOT EXPECT POINTER
						//
						// Output:
						// var a0Name fn.T
						// if a0.Name != nil {
						// 	 tmp := fn.Fn(*a0.Name) // Func requires value input
						//   if tmp != nil {
						//     a0Name = *tmp
						//   }
						// }
						c.Add(
							Var().Add(a0Name()).Add(GenType(rhs)),
							If(a0Selection().Op("!=").Id("nil")).Block(
								Id("tmp").Op(":=").Add(fnCall()),
								If(Id("tmp").Op("!=").Id("nil").Block(
									a0Name().Op("=").Op("*").Id("tmp"),
								)),
							),
						)
					} else {
						// NO ERROR > NOT SLICE > INPUT IS POINTER > OUTPUT IS POINTER > EXPECTS POINTER
						// Output:
						// var a0Name *fn.T
						// if a0.Name != nil {
						// 	 a0Name = fn.Fn(a0.Name)  // Func requires pointer input
						// 	 a0Name = fn.Fn(*a0.Name) // Func requires value input
						// }
						c.Add(
							Var().Add(a0Name()).Add(GenType(rhs)),
							If(a0Selection().Op("!=").Id("nil")).Block(
								a0Name().Op("=").Add(fnCall())),
						)
					}
				} else {
					if expectsPointer {
						// NO ERROR > NOT SLICE > INPUT IS POINTER > OUTPUT IS NOT POINTER > EXPECTS POINTER
						//
						// Output:
						// var a0Name *fn.T
						// if a0.Name != nil {
						// 	 tmp := fn.Fn(a0.Name)  // Func requires pointer input
						//   a0Name = &tmp
						// }
						c.Add(
							Var().Add(a0Name()).Add(GenType(rhs)),
							If(a0Selection().Op("!=").Id("nil")).Block(
								Id("tmp").Op(":=").Add(fnCall()),
								a0Name().Op("=").Op("&").Id("tmp"),
							),
						)

					} else {
						// NO ERROR > NOT SLICE > INPUT IS POINTER > OUTPUT IS NOT POINTER > DOES NOT EXPECT POINTER
						//
						// Output:
						// var a0Name fn.T
						// if a0.Name != nil {
						// 	 a0Name = fn.Fn(a0.Name)  // Func requires pointer input
						// 	 a0Name = fn.Fn(*a0.Name) // Func requires value input
						// }
						c.Add(
							Var().Add(a0Name()).Add(GenType(rhs)),
							If(a0Selection().Op("!=").Id("nil")).Block(
								a0Name().Op("=").Add(fnCall()),
							),
						)
					}
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
						If(Id("each").Op("!=").Id("nil").Block(
							Id("tmp").Op(":=").Add(fnCall()),
							Id("tmp").Op(":=").Id("nil").Block(
								If(Id("tmp").Op("!=").Id("nil")).Block(
									a0Selection().Op("=").Append(a0Selection(), Do(func(s *Statement) {
										if outputIsPointer && !expectsPointer {
											s.Add(Op("*"))
										}
										if !outputIsPointer && expectsPointer {
											s.Add(Op("&"))
										}
									}).Id("tmp")),
								),
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
							If(Id("tmp").Op("!=").Id("nil")).Block(
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

	} else {
		// HAS ERROR.
		if !lhs.IsSlice {
			// IS NOT SLICE
			if inputIsPointer {
				// INPUT IS POINTER
				if outputIsPointer {
					// OUTPUT IS POINTER
					if expectsPointer {
						// HAS ERROR > IS NOT SLICE > INPUT IS POINTER > OUTPUT IS POINTER > EXPECTS POINTER
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
						c.Add(If(a0Selection().Op("!=").Id("nil")).Block(
							List(a0Name(), Id("err")).Op("=").Add(fnCall()),
							b.GenReturnOnError(),
						))
					} else {
						// HAS ERROR > IS NOT SLICE > INPUT IS POINTER > OUTPUT IS POINTER > DOES NOT EXPECT POINTER
						//
						// Output:
						// var a1Name fn.T
						// if a0Name != nil {
						//   tmp, err := fn.Fn(&a0Name)
						//   if err != nil {
						//      return nil, err
						//   }
						//   if tmp != nil  {
						//     a1Name = *tmp
						//   }
						// }
						c.Add(Var().Add(a0Name()).Add(GenType(rhs)))
						c.Add(If(a0Selection().Op("!=").Id("nil")).Block(
							List(Id("tmp"), Id("err")).Op(":=").Add(fnCall()),
							b.GenReturnOnError(),
							If(Id("tmp").Op("!=").Id("nil")).Block(
								a0Name().Op("=").Op("*").Id("tmp"),
							),
						))
					}
				} else {
					// OUTPUT IS NOT POINTER
					if expectsPointer {
						// EXPECTS POINTER
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
							If(a0Selection().Op("!=").Id("nil")).Block(
								List(Id("tmp"), Id("err")).Op(":=").Add(fnCall()),
								b.GenReturnOnError(),
								a0Name().Op("=").Op("&").Id("tmp"),
							),
						)
					} else {
						// DOES NOT EXPECT POINTER
						//
						// Output:
						// var a1Name b.B
						// if a0Name != nil {
						//   a1Name, err = fn.Fn(*a0Name)
						//   if err != nil {
						//      return nil, err
						//   }
						// }
						c.Add(Var().Add(a0Name()).Add(GenType(rhs)))
						c.Add(If(a0Selection().Op("!=").Id("nil")).Block(
							List(a0Name(), Id("err")).Op("=").Add(fnCall()),
							b.GenReturnOnError(),
						))
					}
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
						List(a0Name(), Id("err")).Op(":=").Add(fnCall()),
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
						List(a0Name(), Id("err")).Op(":=").Add(fnCall()),
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
							If(Id("each").Op("!=").Id("nil").Block(
								List(Id("tmp"), Id("err")).Op(":=").Add(fnCall()),
								b.GenReturnOnError(),
								Id("tmp").Op(":=").Id("nil").Block(
									If(Id("tmp").Op("!=").Id("nil")).Block(
										a0Selection().Op("=").Append(a0Selection(), Do(func(s *Statement) {
											if outputIsPointer && !expectsPointer {
												s.Add(Op("*"))
											}
											if !outputIsPointer && expectsPointer {
												s.Add(Op("&"))
											}
										}).Id("tmp")),
									),
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
							If(Id("each").Op("!=").Id("nil").Block(
								List(Id("tmp"), Id("err")).Op(":=").Add(fnCall()),
								b.GenReturnOnError(),
								If(Id("tmp").Op("!=").Id("nil")).Block(
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
							List(Id("tmp"), Id("err")).Op(":=").Add(fnCall()),
							b.GenReturnOnError(),
							If(Id("tmp").Op("!=").Id("nil")).Block(
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
							List(Id("tmp"), Id("err")).Op(":=").Add(fnCall()),
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

// validateFunctionSignatureMatch ensures that the conversion from input to
// output is allowed.
// Function can receive value/pointer.
// Function must return a value/pointer with optional error.
// Function must accept the input signature of the type.
// Input can be one or many.
// Function must accept struct only, if the input is many, it will be
// operated at elem level.
func (b *FuncBuilder) validateFunctionSignatureMatch(fn *mapper.Func, lhs, rhs *mapper.Type) {
	var (
		in                  = fn.From.Type
		out                 = fn.To.Type
		pointerToNonPointer = lhs.IsPointer && !rhs.IsPointer
		//isMany              = fn.From.Type.IsSlice || fn.From.Variadic
	)

	// Slice A might not equal A
	// []A != A
	if !in.Equal(lhs) {
		// But internally, the type matches. This is allowed because we may have a
		// private mapper that maps A.
		// A == A
		if in.Type != lhs.Type {
			panic(ErrMismatchType(in, lhs))
		}
	}

	if !out.Equal(rhs) {
		if out.Type != rhs.Type {
			panic(ErrMismatchType(out, rhs))
		}
	}
	if pointerToNonPointer {
		panic(fmt.Sprintf("mapper: func cannot return non-pointer for value input: %s", fn.Signature()))
	}

	//if isMany {
	//panic(fmt.Sprintf("mapper: func input must be struct: %s, got %s", fn.Signature(), lhs.Signature()))
	//}
}

func ErrMismatchType(lhs, rhs *mapper.Type) error {
	return fmt.Errorf(`mapper: signature does not match: %s to %s`,
		lhs.Signature(),
		rhs.Signature(),
	)
}
