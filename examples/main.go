package main

import (
	"fmt"
	"go/types"
	"log"
	"strconv"

	"github.com/alextanhongpin/mapper"
	"github.com/alextanhongpin/mapper/examples/bar"
	"github.com/alextanhongpin/mapper/examples/foo"
	"github.com/dave/jennifer/jen"
	. "github.com/dave/jennifer/jen"
	"github.com/google/uuid"
)

const Generator = "mapper"

type Task struct {
	Name string
}

func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

type Foo struct {
	CustomID string `map:"ExternalID,ParseUUID"`
	FakeAge  int    `map:"Age"`
	name     string
	Task     Task
	id       string
	//Remarks string // Fail with extra fields now.
	// CustomName `mapper:"YourName"`
	// CustomMapper `mapper:"github.com/yourorganization/yourpackage/struct.Method"`
	// CustomInterface `mapper:"github.com/yourorganization/yourpackage/interface.Method"`
	// CustomFunction`mapper:"github.com/yourorganization/yourpackage.funcName"`
}

func (f Foo) ID() (uuid.UUID, error) {
	return uuid.Parse(f.id)
}

//func (f Foo) Name(ctx context.Context) string { // Panics, since it accepts arguments.
func (f Foo) Name() string {
	return f.name
}

func CustomConverter(a string) int {
	i, _ := strconv.Atoi(a)
	return i
}

//go:generate go run main.go -type Converter
type Converter interface {
	ConvertNameless(Foo) (Bar, error) // Accepts err.
	Convert(a Foo) (Bar, error)       // Accepts err.
	ConvertImport(f foo.Foo) (b bar.Bar, err error)
	// Pointers not accepted
	//ConvertReturnPointer(a Foo) (*Bar, error) // Accepts err.
	//ConvertImportPointer(f *foo.Foo) (b *bar.Bar, err error)
	//ConvertWithContext(ctx context.Context, foo Foo) (Bar)
	ConvertSlice(a []Foo) ([]Bar, error) // Accepts err.
	ConvertSliceWithoutErrors(a []A) []B // Accepts err.
}

type A struct {
	Name string
}

type B struct {
	Name string
}

type Bar struct {
	ID         uuid.UUID
	Name       string
	RealAge    int `json:"age" map:"Age"`
	Task       Task
	ExternalID uuid.UUID
}

func main() {
	if err := mapper.New(generateConverterFromFields); err != nil {
		log.Fatalln(err)
	}
}

func generateConverterFromFields(opt mapper.Option) error {
	var (
		pkgName  = opt.PkgName
		pkgPath  = opt.PkgPath
		typeName = opt.TypeName
		typ      = opt.Type
		out      = opt.Out
	)

	f := NewFile(pkgName) // e.g. main
	f.PackageComment(fmt.Sprintf("Code generated by %s, DO NOT EDIT.", Generator))

	// TODO: Inject dependency injection.
	generateConverter(f, typeName)
	generateConverterConstructor(f, typeName)

	privateMethods := make(map[string]bool)
	for _, converter := range typ.InterfaceMethods {
		if converter.From.Type.IsPointer || converter.To.Type.IsPointer {
			panic("mapper: conversion to and/or from struct pointer is not allowed")
		}
		// Cannot convert pointers.
		// Cannot convert if error type is not fulfilled
		// Cannot convert struct to slice/or vice versa.
		if privateMethods[converter.NormalizedSignature()] {
			continue
		}
		generatePrivateMethod(f, pkgPath, typeName, converter)
		privateMethods[converter.NormalizedSignature()] = true
	}

	for _, converter := range typ.InterfaceMethods {
		// There is a normalized private method mapper, use it.
		if privateMethods[converter.NormalizedSignature()] {
			generateUsePrivateMapper(f, pkgPath, typeName, converter)
		} else {
			generateConvertMethod(f, pkgPath, typeName, converter)
		}
	}

	return f.Save(out) // e.g. main_gen.go
}

func generateConverter(f *jen.File, typeName string) {
	// Output:
	//type Converter struct {
	//}

	f.Type().Id(typeName).Struct().Line()
}

func generateConverterConstructor(f *jen.File, typeName string) {
	// Output:
	// func NewConverter() *Converter {
	//   return &Converter{
	//   }
	// }

	f.Func().Id(fmt.Sprintf("New%s", typeName)).Params().Op("*").Id(typeName).Block(
		Return(Op("&").Id(typeName).Values(Dict{})),
	).Line()
}

func generateConvertMethod(f *jen.File, pkgPath, typeName string, fn mapper.Func) {
	// Output:
	// func (c *Converter) Convert(a A) (B, error) {
	//   return B{
	//     ID: a.ID,
	//     Name: a.Name,
	//   }, nil
	// }

	from, to := fn.From, fn.To

	dict := Dict{}

	type mapperFunc struct {
		Fn *mapper.Func
		In mapper.StructField
	}
	var funcsWithError []mapperFunc
	var methodsWithError []mapper.Func
	for key, t := range to.Type.StructFields {

		// Check if there is a field mapping.
		f, ok := from.Type.StructFields[key]
		if ok {
			fmt.Println(fn.Name, f)

			// `map:"CustomField,CustomFunc"`
			if tag := f.Tag; tag != nil && tag.HasFunc() {
				// Load the function.
				pkg := mapper.LoadPackage(pkgPath)
				obj := mapper.LookupType(pkg, tag.Func)
				if obj == nil {
					panic(fmt.Sprintf("mapper: func not found: %s", tag.Func))
				}

				fnType, ok := obj.(*types.Func)
				if !ok {
					panic(fmt.Sprintf("mapper: not a func: %s", tag.Func))
				}

				fn := mapper.ExtractFunc(fnType)

				if fn.From.Type.Type != f.Type.Type {
					panic(fmt.Sprintf("mapper: input signature does not match: %v != %v", fn.From.Type.Type, f.Type.Type))
				}

				if fn.To.Type.Type != t.Type.Type {
					panic(fmt.Sprintf("mapper: output signature does not match: %v != %v", fn.To.Type.Type, t.Type.Type))
				}

				if fn.Error != nil {
					funcsWithError = append(funcsWithError, mapperFunc{
						Fn: fn,
						In: f,
					})
					// Name: aName,
					dict[Id(t.Name)] = Id(from.Name + f.Name)
					continue
				}

				// Name: ParseUUID(a.Name)
				dict[Id(t.Name)] = Id(fn.Name).Call(Id(from.Name).Dot(f.Name))

				continue
			}

			// Name: a.Name
			dict[Id(t.Name)] = Id(from.Name).Dot(f.Name)
			continue
		}

		// Check if there is a method with the name that returns the same signature.
		var found bool
		for _, m := range from.Type.StructMethods {
			// The name of the method matches the name of field, e.g. to.Age: from.Age()
			if m.Name == key {
				if m.To.Type.Type != t.Type.Type {
					panic("method signature found, but types are different")
				}
				if m.From != nil {
					panic("method must not accept any arguments")
				}
				found = true

				// TODO: If this has error, handle the return errors.
				if m.Error != nil {
					// Name: aName,
					dict[Id(t.Name)] = Id(from.Name + m.Name)
					methodsWithError = append(methodsWithError, m)
				} else {
					// Name: a.Name()
					dict[Id(t.Name)] = Id(from.Name).Dot(m.Name).Call()
				}
				break
			}
		}
		if !found {
			panic("method signature not found")
		}
	}

	var toReturnPtr, toPtr string
	if to.Type.IsPointer {
		toPtr = "&"
		toReturnPtr = "*"
	}

	var fromPtr string
	if from.Type.IsPointer {
		fromPtr = "*"
	}

	var returnType, returnVar Code
	if fn.Error != nil {
		returnType = Parens(List(
			Op(toReturnPtr).Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type),
			Id("error"),
		))
		returnVar = List(
			Op(toPtr).Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(dict),
			Id("nil"),
		)
	} else {
		returnType = Op(toReturnPtr).Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type)
		returnVar = Op(toPtr).Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(dict)
	}

	// For each methods that has error as the second return value,
	// initialize them and return it.
	var checkErrors []Code
	for _, fn := range methodsWithError {
		// aName, err := a.Name()
		// if err != nil {
		//	return &Bar{}, err
		// }
		checkErrors = append(checkErrors, List(Id(from.Name+fn.Name), Id("err")).Op(":=").Id(from.Name).Dot(fn.Name).Call())
		checkErrors = append(checkErrors,
			If(Id("err").Op("!=").Id("nil").Block(
				Return(
					List(
						Op(toPtr).Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(Dict{}),
						Id("err"),
					),
				))).Line())
	}

	for _, fn := range funcsWithError {
		// aName, err := CustomFunc(a.Name)
		// if err != nil {
		//	return &Bar{}, err
		// }
		checkErrors = append(checkErrors, List(Id(from.Name+fn.In.Name), Id("err")).Op(":=").Id(fn.Fn.Name).Call(Id(from.Name).Dot(fn.In.Name)))
		checkErrors = append(checkErrors,
			If(Id("err").Op("!=").Id("nil").Block(
				Return(
					List(
						Op(toPtr).Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(Dict{}),
						Id("err"),
					),
				))).Line())
	}

	f.Func().Params(Id("c").Op("*").Id(typeName)). // (c *Converter)
							Id(fn.Name).Params(Id(from.Name).Op(fromPtr).Qual(relativeTo(pkgPath, from.Type.PkgPath), from.Type.Type)). // Convert(a A)
							Add(returnType).Block(
		append(checkErrors, Return(returnVar))...,
	).Line()
}

// generatePrivateMethod generates the most basic type for the given function
// to be reused for slice mapping.
func generatePrivateMethod(f *jen.File, pkgPath, typeName string, fn mapper.Func) {
	// Output:
	// func (c *Converter) mapMainAToMainB(a A) (B, error) {
	//   return B{
	//     ID: a.ID,
	//     Name: a.Name,
	//   }, nil
	// }

	fnName := fn.NormalizedName()
	from, to := fn.From, fn.To

	dict := Dict{}

	type mapperFunc struct {
		Fn *mapper.Func
		In mapper.StructField
	}
	var funcsWithError []mapperFunc
	var methodsWithError []mapper.Func
	for key, t := range to.Type.StructFields {

		// Check if there is a field mapping.
		f, ok := from.Type.StructFields[key]
		if ok {
			fmt.Println(fn.Name, f)

			// `map:"CustomField,CustomFunc"`
			if tag := f.Tag; tag != nil && tag.HasFunc() {
				// Load the function.
				pkg := mapper.LoadPackage(pkgPath)
				obj := mapper.LookupType(pkg, tag.Func)
				if obj == nil {
					panic(fmt.Sprintf("mapper: func not found: %s", tag.Func))
				}

				fnType, ok := obj.(*types.Func)
				if !ok {
					panic(fmt.Sprintf("mapper: not a func: %s", tag.Func))
				}

				fn := mapper.ExtractFunc(fnType)

				if fn.From.Type.Type != f.Type.Type {
					panic(fmt.Sprintf("mapper: input signature does not match: %v != %v", fn.From.Type.Type, f.Type.Type))
				}

				if fn.To.Type.Type != t.Type.Type {
					panic(fmt.Sprintf("mapper: output signature does not match: %v != %v", fn.To.Type.Type, t.Type.Type))
				}

				if fn.Error != nil {
					funcsWithError = append(funcsWithError, mapperFunc{
						Fn: fn,
						In: f,
					})
					// Name: aName,
					dict[Id(t.Name)] = Id(from.Name + f.Name)
					continue
				}

				// Name: ParseUUID(a.Name)
				dict[Id(t.Name)] = Id(fn.Name).Call(Id(from.Name).Dot(f.Name))

				continue
			}

			// Name: a.Name
			dict[Id(t.Name)] = Id(from.Name).Dot(f.Name)
			continue
		}

		// Check if there is a method with the name that returns the same signature.
		var found bool
		for _, m := range from.Type.StructMethods {
			// The name of the method matches the name of field, e.g. to.Age: from.Age()
			if m.Name == key {
				if m.To.Type.Type != t.Type.Type {
					panic("method signature found, but types are different")
				}
				if m.From != nil {
					panic("method must not accept any arguments")
				}
				found = true

				// TODO: If this has error, handle the return errors.
				if m.Error != nil {
					// Name: aName,
					dict[Id(t.Name)] = Id(from.Name + m.Name)
					methodsWithError = append(methodsWithError, m)
				} else {
					// Name: a.Name()
					dict[Id(t.Name)] = Id(from.Name).Dot(m.Name).Call()
				}
				break
			}
		}
		if !found {
			panic("method signature not found")
		}
	}

	// For each methods that has error as the second return value,
	// initialize them and return it.
	genMethodsWithError := func(g *Group) {
		for _, fn := range methodsWithError {
			// aName, err := a.Name()
			// if err != nil {
			//	return Bar{}, err
			// }
			g.Add(List(Id(from.Name+fn.Name), Id("err")).Op(":=").Id(from.Name).Dot(fn.Name).Call())
			g.Add(
				If(Id("err").Op("!=").Id("nil").Block(
					Return(
						List(
							Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(Dict{}),
							Id("err"),
						),
					)),
				).Line(),
			)
		}
	}

	genFuncsWithError := func(g *Group) {
		for _, fn := range funcsWithError {
			// aName, err := CustomFunc(a.Name)
			// if err != nil {
			//	return Bar{}, err
			// }
			g.Add(List(Id(from.Name+fn.In.Name), Id("err")).Op(":=").Id(fn.Fn.Name).Call(Id(from.Name).Dot(fn.In.Name)))
			g.Add(If(Id("err").Op("!=").Id("nil").Block(
				Return(
					List(
						Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(Dict{}),
						Id("err"),
					),
				)),
			).Line())
		}
	}

	genReturnType := func(s *Statement) {
		if fn.Error != nil {
			// (Bar, nil)
			s.Add(Parens(List(
				Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type),
				Id("error"),
			)))
		} else {
			// (Bar)
			s.Add(Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type))
		}
	}

	f.Func().
		Params(Id("c").Op("*").Id(typeName)). // (c *Converter)
		Id(fnName).                           // mapMainAToMainB
		Params(
			Id(from.Name).Qual(relativeTo(pkgPath, from.Type.PkgPath), from.Type.Type),
		). // (a A)
		Do(genReturnType).
		BlockFunc(func(g *Group) {
			// aName, err := a.Name()
			// if err != nil {
			//	return Bar{}, err
			// }
			genMethodsWithError(g)

			// aName, err := CustomFunc(a.Name)
			// if err != nil {
			//	return Bar{}, err
			// }
			genFuncsWithError(g)

			g.Add(ReturnFunc(func(g *Group) {
				if fn.Error != nil {
					// return Bar{}, nil
					g.Add(
						List(
							Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(dict),
							Id("nil"),
						),
					)
				} else {
					// return Bar{}
					g.Add(Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(dict))
				}
			}))
		}).Line()
}

func generateUsePrivateMapper(f *jen.File, pkgPath, typeName string, fn mapper.Func) {
	// Output:
	// func (c *Converter) Convert(a A) (B, error) {
	//   *body*
	// }

	// Where body is one of these scenario.
	// No error
	// return c.mapMainAToMainB(main.A)

	// Input pointer.
	// return c.mapMainAToMainB(*main.A)

	// Output pointer.
	// return *c.mapMainAToMainB(main.A)

	// Input/Output pointer.
	// return *c.mapMainAToMainB(*main.A)

	// With error and pointers.
	// mainB, err := c.mapMaihAToMainB(&main.A)
	// return *mainB, err

	from, to := fn.From, fn.To
	if from.Type.IsSlice != to.Type.IsSlice {
		panic("mapper: slice to no-slice and vice versa is not allowed")
	}
	isSlice := from.Type.IsSlice
	inType := Qual(relativeTo(pkgPath, from.Type.PkgPath), from.Type.Type)
	outType := Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type)

	genInputType := func(g *Group) {
		g.Add(
			Id(from.Name).Do(func(s *Statement) {
				if from.Type.IsSlice {
					s.Add(Index())
				}
			}).Op(pointerOp(from.Type, "*")).Add(inType),
		)
	}

	genReturnType := func(s *Statement) {
		if fn.Error != nil {
			// (*Bar, nil)
			s.Add(Parens(ListFunc(func(g *Group) {
				g.Add(
					Op(pointerOp(to.Type, "*")).Do(func(rs *Statement) {
						if to.Type.IsSlice {
							rs.Add(Index())
						}
					}).Add(outType))
				g.Add(Id("error"))
			})))
		} else {
			// (*Bar)
			s.Add(
				Op(pointerOp(to.Type, "*")).Do(func(s *Statement) {
					if to.Type.IsSlice {
						s.Add(Index())
					}
				}).Add(outType),
			)
		}
	}

	f.Func().
		Params(Id("c").Op("*").Id(typeName)). // (c *Converter)
		Id(fn.Name).ParamsFunc(genInputType). // Convert(a *A)
		Do(genReturnType).                    // (*B, error)
		BlockFunc(func(g *Group) {
			if isSlice {
				// var err error
				// res := make([]B, len(a))
				// for i, s := range a {
				//   res[i], err = c.mapMainAToMainB(s)
				//   if err != nil { return err }
				// }
				// return res, nil
				if fn.Error != nil {
					g.Add(Var().Id("err").Id("error"))
				}
				g.Add(Id("res").Op(":=").Make(List(Index().Add(outType), Len(Id(from.Name)))))
				g.Add(For(List(Id("i"), Id("s")).Op(":=").Range().Id(from.Name)).BlockFunc(func(g *Group) {
					if fn.Error != nil {
						g.Add(List(
							Id("res").Index(Id("i")),
							Id("err"),
						).Op("=").Id("c").Dot(fn.NormalizedName()).Call(Id("s")))
						g.Add(If(Id("err").Op("!=").Id("nil")).Block(ReturnFunc(func(g *Group) {
							if isSlice {
								g.Add(List(Id("nil"), Id("err")))
							} else {
								g.Add(List(outType.Clone().Values(), Id("err")))
							}
						})))
					} else {
						g.Add(Id("res").Index(Id("i")).Op("=").Id("c").Dot(fn.NormalizedName()).Call(Id("s")))
					}
				}))

				if fn.Error != nil {
					// return res, nil
					g.Add(Return(List(Id("res"), Id("nil"))))
				} else {
					// return res
					g.Add(Return(Id("res")))
				}
			} else {
				// Ignore pointers.
				g.Add(Return(Id("c").Dot(fn.NormalizedName()).Call(Id(from.Name)))) // return c.mapMainAToMainB(a)
			}
		}).Line()
}

func relativeTo(pkgPath, fieldPkgPath string) string {
	if pkgPath == fieldPkgPath {
		return ""
	}
	return fieldPkgPath
}

func pointerOp(m *mapper.Type, op string) string {
	if !m.IsPointer {
		return ""
	}
	return op
}
