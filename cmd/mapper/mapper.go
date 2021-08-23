package main

import (
	"fmt"
	"go/types"
	"log"
	"sort"

	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
	. "github.com/dave/jennifer/jen"
)

const GeneratorName = "github.com/alextanhongpin/mapper"

type C []Code

func (c *C) Add(code ...Code) *C {
	*c = append(*c, code...)
	return c
}

func main() {
	if err := mapper.New(func(opt mapper.Option) error {
		gen := NewGenerator(opt)
		return gen.Generate()
	}); err != nil {
		log.Fatalln(err)
	}
}

type Generator struct {
	opt              mapper.Option
	uses             map[string]mapper.Type
	mappers          map[string]bool
	hasErrorByMapper map[string]bool
}

func NewGenerator(opt mapper.Option) *Generator {
	return &Generator{
		opt:              opt,
		uses:             make(map[string]mapper.Type),
		mappers:          make(map[string]bool),
		hasErrorByMapper: make(map[string]bool),
	}
}

type FieldResolver interface {
	Name() string
	Var() *Statement
	Selection() *Statement
	Lhs() interface{}
	Rhs() mapper.StructField
	IsField() bool
	IsMethod() bool
}

type fieldResolver struct {
	method *mapper.Func
	field  *mapper.StructField
	name   string
	rhs    mapper.StructField
}

func NewMethodFieldResolver(method *mapper.Func, name string, rhs mapper.StructField) *fieldResolver {
	return &fieldResolver{
		method: method,
		name:   argsWithIndex(name, 0),
		rhs:    rhs,
	}
}

func NewStructFieldResolver(field *mapper.StructField, name string, rhs mapper.StructField) *fieldResolver {
	return &fieldResolver{
		field: field,
		name:  argsWithIndex(name, 0),
		rhs:   rhs,
	}
}

func (f fieldResolver) SetName(name string) {
	f.name = argsWithIndex(name, 0)
}

func (f fieldResolver) Name() string {
	return f.name
}

func (f fieldResolver) Var() *Statement {
	if f.method != nil {
		return Id(f.name + f.method.Name).Clone()
	}
	if f.field != nil {
		return Id(f.name + f.field.Name).Clone()
	}
	return nil
}

func (f fieldResolver) Selection() *Statement {
	if f.method != nil {
		return Id(f.name).Dot(f.method.Name).Call().Clone()
	}
	if f.field != nil {
		return Id(f.name).Dot(f.field.Name).Clone()
	}
	return nil
}

func (f fieldResolver) Rhs() mapper.StructField {
	return f.rhs
}

func (f fieldResolver) Lhs() interface{} {
	if f.IsField() {
		return f.field
	}

	if f.IsMethod() {
		return f.method
	}

	panic("mapper: resolver must be field or method")
}

func (f fieldResolver) IsField() bool {
	return f.field != nil && f.method == nil
}

func (f fieldResolver) IsMethod() bool {
	return f.method != nil && f.field == nil
}

func (g *Generator) Generate() error {
	var (
		pkgPath = g.opt.PkgPath
		pkgName = g.opt.PkgName
		typ     = g.opt.Type
		out     = g.opt.Out
	)

	// Since a package path basename might not be the same as the package name,
	// This allows us to use Qual and exclude imports from the same package.
	f := NewFilePathName(pkgPath, pkgName)
	f.PackageComment(fmt.Sprintf("Code generated by %s, DO NOT EDIT.", GeneratorName))

	// Cache first so that we can re-use later.
	var keys []string
	for key, method := range typ.InterfaceMethods {
		g.mappers[method.NormalizedSignature()] = false
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var stmts []*Statement
	for _, key := range keys {
		method := typ.InterfaceMethods[key]
		g.validateToAndFromStruct(method)
		if g.mappers[method.NormalizedSignature()] {
			continue
		}
		stmt := g.genPrivateMethod(Null(), method)
		stmts = append(stmts, stmt)
		g.mappers[method.NormalizedSignature()] = true
	}

	// Generate the struct and constructor before the method declarations.
	g.genStruct(f)
	g.genConstructor(f)

	for _, stmt := range stmts {
		f.Add(stmt)
	}

	for _, key := range keys {
		method := typ.InterfaceMethods[key]
		if !g.mappers[method.NormalizedSignature()] {
			panic("mapper: method not found")
		}
		g.genPublicMethod(f, method)
	}

	return f.Save(out) // e.g. main_gen.go
}

func (g *Generator) genStruct(f *jen.File) {
	// Output:
	// type Converter struct {
	//   customInterface interfacepkg.CustomInterface
	//   customStruct    *structpkg.CustomStruct
	// }

	f.Type().Id(g.opt.TypeName).StructFunc(func(group *Group) {
		for typeName, use := range g.uses {
			group.Add(Id(typeName), Do(func(s *Statement) {
				if use.IsStruct {
					s.Add(Op("*"))
				}
			}).Qual(use.PkgPath, use.Type))
		}
	}).Line()
}

func (g *Generator) genConstructor(f *jen.File) {
	// Output:
	// func NewConverter(customStruct *structpkg.CustomStruct, customInterface interfacepkg.CustomInterface) *Converter {
	//   return &Converter{
	//     structpkgCustomStruct: customStruct,
	//     interfacepkgCustomInterface: customInterface,
	//   }
	// }

	typeName := g.opt.TypeName

	f.Func().Id(fmt.Sprintf("New%s", typeName)).ParamsFunc(func(group *Group) {
		for structName, use := range g.uses {
			group.Add(Id(structName), Do(func(s *Statement) {
				if use.IsStruct {
					s.Add(Op("*"))
				}
			}).Qual(use.PkgPath, use.Type))
		}
	}).Op("*").Id(typeName).Block(
		Return(Op("&").Id(typeName).ValuesFunc(func(group *Group) {
			dict := make(Dict)
			for structName := range g.uses {
				dict[Id(structName)] = Id(structName)
			}
			group.Add(dict)
		})),
	).Line()
}

type mapperFunc struct {
	Fn         *mapper.Func
	In         mapper.StructField
	structName string
	callerID   *Statement // customfuncpkg.CustomFunc | c.customstructpkg.structname.StructMethod | c.custominterfacepkg.interfacename.InterfaceMethod
	resolver   FieldResolver
}

func (m mapperFunc) CallerID() *Statement {
	return m.callerID.Clone()
}

// genPrivateMethod generates the most basic, struct A to struct B conversion
// without pointers, slice etc.
func (g *Generator) genPrivateMethod(f *jen.Statement, fn mapper.Func) *jen.Statement {
	var (
		pkgPath  = g.opt.PkgPath
		typeName = g.opt.TypeName
		fnName   = fn.NormalizedName()
		from     = fn.From
		to       = fn.To
	)

	var mappersWithError []mapperFunc
	var pointers, methodsWithError C

	genReturnOnError := func() *Statement {
		// Output:
		// if err != nil {
		//   return Bar{}, err
		// }
		return If(Id("err").Op("!=").Id("nil").Block(
			Return(
				List(
					genType(to.Type).Values(Dict{}),
					Id("err"),
				),
			)),
		).Clone()
	}

	var keys []string
	for key := range to.Type.StructFields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var resolvers []FieldResolver
	for _, key := range keys {
		to := to.Type.StructFields[key]
		// If LHS field matche the RHS field ...
		field, ok := from.Type.StructFields[key]
		if ok {
			resolver := NewStructFieldResolver(&field, from.Name, to)
			resolvers = append(resolvers, resolver)
			continue
		}

		// If LHS has a method that maps to RHS field,
		// e.g. RHS.Name == LHS.Name()
		// LHS method can also return error as the second argument.
		if method, ok := from.Type.StructMethods[key]; ok {
			resolver := NewMethodFieldResolver(&method, from.Name, to)
			resolvers = append(resolvers, resolver)
			continue
		}
		panic("mapper: method signature not found")
	}

	dict := make(Dict)
	for _, resolver := range resolvers {
		var (
			rhs = resolver.Rhs().Type
			// Argument name has a `0` to indicate it's position, which is useful to
			// avoid conflict in naming.
			a0Name      = resolver.Var
			a0Selection = resolver.Selection

			// The return type field name.
			bName = resolver.Rhs().Name
		)

		// Resolves method. Note that tag function transformation does not apply to
		// methods.
		if resolver.IsMethod() {
			method := resolver.Lhs().(*mapper.Func)
			if method.Error != nil {
				// Output:
				// a0Name, err := a0.Name()
				// if err != nil {
				//   return B{}, err
				// }
				methodsWithError.Add(
					List(a0Name(), Id("err")).Op(":=").Add(a0Selection()),
					genReturnOnError().Line(),
				)

				// Output:
				// name: a0Name
				dict[Id(bName)] = a0Name()
			} else {
				// Output:
				// name: a0.Name()
				dict[Id(bName)] = a0Selection()
			}
			continue
		}

		var (
			left = resolver.Lhs().(*mapper.StructField)
			lhs  = left.Type
		)

		// Check if there is a tag transformation.
		if tag := left.Tag; tag != nil && tag.HasFunc() {
			// The tag defines a custom function, TransformationFunc that can be used to
			// map LHS field to RHS.
			if tag.IsFunc() {
				fn := g.loadTagFunction(tag)
				g.validateFunctionSignatureMatch(fn, lhs, rhs)

				// If the TransformationFunc returns error, it needs to be handled and
				// returned early.
				if fn.Error != nil {
					mappersWithError = append(mappersWithError, mapperFunc{
						Fn:       fn,
						In:       *left,
						resolver: resolver,
						// funcpkg.CustomFunc
						callerID: Qual(fn.PkgPath, fn.Name),
					})
					// Name: aName,
					dict[Id(bName)] = a0Name()
					continue
				}

				// No errors.
				// RHS field is a pointer.
				// LHS needs to be converted to a pointer too.
				g.validatePointerConversion(lhs, rhs)
				if rhs.IsPointer {
					// customfuncpkg.CustomFunc(a0.Name)
					callerID := Qual(fn.PkgPath, fn.Name)
					if lhs.IsPointer {
						// Output:
						// var a0Name *outpkg.OutType
						// if a0.Name != nil {
						//   res := customfuncpkg.CustomFunc(*a0.Name)
						//   a0Name = &res
						// }
						pointers.Add(
							Var().Add(a0Name()).Op("*").Add(genType(rhs)),
							If(a0Selection()).Block(
								Id("tmp").Op(":=").Add(callerID.Call(Op("*").Add(a0Selection()))),
								a0Name().Op("=").Op("&").Id("tmp"),
							),
						)

						// Output:
						// Name: a0Name
						dict[Id(bName)] = a0Name()
					} else {
						// Output:
						// a0Name := customfuncpkg.CustomFunc(a0.Name)
						pointers.Add(a0Name().Op(":=").Add(callerID.Call(a0Selection())))

						// Output:
						// Name: &a0Name
						dict[Id(bName)] = Op("&").Add(a0Name())
					}
					continue
				}

				// name: customfuncpkg.CustomFunc(a.Name)
				dict[Id(bName)] = Qual(fn.PkgPath, fn.Name).Call(a0Selection())
				continue
			}

			// The tag loads a custom struct or interface method.
			if tag.IsMethod() {
				fieldPkgPath := pkgPath
				if tag.IsImported() {
					fieldPkgPath = tag.PkgPath
				}

				// Load the function.
				pkg := mapper.LoadPackage(fieldPkgPath)
				obj := mapper.LookupType(pkg, tag.TypeName)
				if obj == nil {
					panic(fmt.Sprintf("mapper: type not found: %s", tag.TypeName))
				}

				if _, ok := obj.Type().(*types.Named); !ok {
					panic("mapper: not a named type")
				}

				typ := mapper.NewType(obj.Type())
				var method mapper.Func
				switch {
				case typ.IsInterface:
					method = typ.InterfaceMethods[tag.Func]
				case typ.IsStruct:
					method = typ.StructMethods[tag.Func]
				default:
					panic("mapper: tag is invalid")
				}

				g.validateFunctionSignatureMatch(&method, lhs, rhs)

				// To avoid different packages having same struct name, prefix the
				// struct name with the package name.
				structPackageName := tag.Pkg + tag.TypeName
				if tag.Pkg == "" {
					structPackageName = mapper.LowerFirst(structPackageName)
				}
				g.uses[structPackageName] = *typ

				if method.Error != nil {
					mappersWithError = append(mappersWithError, mapperFunc{
						Fn:         &method,
						In:         *left,
						structName: structPackageName,
						resolver:   resolver,
						// c.structpkgCustomStruct.CustomMethod(a0.Name)
						callerID: g.genShortName().Dot(structPackageName).Dot(method.Name).Call(a0Selection()),
					})
					// Name: aName,
					dict[Id(bName)] = a0Name()
					continue
				}

				g.validatePointerConversion(lhs, rhs)
				if rhs.IsPointer {
					// c.interfacePkgInterface.CustomMethod
					callerID := Add(g.genShortName()).Dot(structPackageName).Dot(method.Name)
					if lhs.IsPointer {
						// Output:
						// var a0Name *outpkg.OutType
						// if a0.Name != nil {
						//   res := c.interfacePkgInterface.CustomMethod(*a.Name)
						//   a0Name = &res
						// }
						pointers.Add(
							Var().Add(a0Name()).Op("*").Add(genType(rhs)),
							If(a0Selection()).Block(
								Id("res").Op(":=").Add(callerID.Call(Op("*").Add(a0Selection()))),
								a0Name().Op("=").Op("&").Id("res"),
							),
						)
						// a0Name: a0Name
						dict[Id(bName)] = a0Name()
					} else {
						// a0Name := c.interfacePkgInterface.CustomMethod(a.Name)
						// Name: &a0Name
						pointers.Add(a0Name().Op(":=").Add(callerID.Call(a0Selection())))
						dict[Id(bName)] = Op("&").Add(a0Name())
					}
					continue

				}
				// Name: c.interfacePkgInterface.CustomMethod(a.Name)
				dict[Id(bName)] = g.genShortName().Dot(structPackageName).Dot(method.Name).Call(a0Selection())
				continue
			}
		} // End of custom tag function.

		// LHS field does not match RHS field for some reason.
		// This could be a struct to struct conversion that requires another
		// internally implemented mapper.
		if !lhs.Equal(rhs) {
			// Check if there is a private mapper with the signature that accepts LHS
			// and returns RHS .
			methodSignature := buildFnSignature(lhs, rhs)
			_, hasSignature := g.mappers[methodSignature]
			if !hasSignature {
				panic(ErrConversion(lhs, rhs))
			}

			var method *mapper.Func
			for _, met := range g.opt.Type.InterfaceMethods {
				if met.NormalizedSignature() == methodSignature {
					method = &met
					break
				}
			}
			if method == nil {
				panic(fmt.Sprintf("mapper: no signature found %s", methodSignature))
			}

			// If the mapper method returns an error, or if it is actually a slice
			// to slice conversion ...
			if method.Error != nil || rhs.IsSlice {
				mappersWithError = append(mappersWithError, mapperFunc{
					Fn:       method,
					In:       *left,
					resolver: resolver,
					// c.mapAToB
					callerID: g.genShortName().Dot(method.NormalizedName()),
				})
				// Output:
				// Name: a0Name,
				dict[Id(bName)] = a0Name()
				continue
			}

			g.validatePointerConversion(lhs, rhs)

			// RHS is pointer. LHS needs to be converted to a pointer too.
			if rhs.IsPointer {
				// c.mapAtoB(a.Name)
				callerID := g.genShortName().Dot(method.NormalizedName())

				// LHS is a pointer, ensure there is no nil pointer conversion.
				if lhs.IsPointer {
					// Output:
					// var a0Name *outpkg.OutType
					// if a0.Name != nil {
					//   res := c.mapAtoB(*a.Name)
					//   a0Name = &res
					// }
					pointers.Add(
						Var().Add(a0Name()).Op("*").Add(genType(rhs)),
						If(a0Selection().Op("!=").Id("nil")).Block(
							Id("res").Op(":=").Add(callerID.Call(Op("*").Add(a0Selection()))),
							a0Name().Op("=").Op("&").Id("res"),
						),
					)
					// Output:
					// a0Name: a0Name
					dict[Id(bName)] = a0Name()
					continue
				}
				// Output:
				// a0Name := c.mapAtoB(a.Name)
				pointers.Add(
					a0Name().Op(":=").Add(callerID.Call(a0Selection())),
				)

				// Output:
				// Name: &a0Name
				dict[Id(bName)] = Op("&").Add(a0Name())
				continue
			}

			// Name: c.mapAtoB(a.Name)
			dict[Id(bName)] = g.genShortName().Dot(method.NormalizedName()).Call(a0Selection())
			continue
		}

		// Output:
		// Name: a.Name
		dict[Id(bName)] = a0Selection()
		continue
	}

	genPrivateMapperMethodsWithError := func(g *Group) {
		for _, fn := range mappersWithError {

			var (
				lhs         = fn.In.Type
				rhs         = fn.Fn.To.Type
				method      = fn.Fn
				a0Name      = fn.resolver.Var
				a0Selection = fn.resolver.Selection
			)
			// []*B
			returnType := Do(func(s *Statement) {
				if lhs.IsSlice {
					s.Add(Index())
				}
				if lhs.IsPointer {
					s.Add(Op("*"))
				}
			}).Add(genType(rhs))

			if lhs.IsSlice {
				// aName := make([]B, len(a.Name))
				g.Add(a0Name().Op(":=").Make(returnType, Len(a0Selection())))

				// The mapper method has error, handle it.
				if method.Error != nil {
					// Output:
					// aName := make([]B, len(a.Name))
					// for i, each := range a.Name {
					//   var err error
					//   aName[i], err = callerFunc(a.Name)
					//   if err != nil {
					//     return Bar{}, err
					//   }
					// }
					g.Add(For(List(Id("i"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
						Var().Id("err").Id("error"),
						List(a0Name().Index(Id("i")), Id("err")).Op("=").Add(fn.CallerID()).Call(Id("each")),
						genReturnOnError(),
					).Line())
					continue
				}
				// Output:
				// aName := make([]B, len(a.Name))
				// for i, each := range a.Name {
				//   aName[i] := callerFn(a.Name)
				// }
				g.Add(For(List(Id("i"), Id("each")).Op(":=").Range().Add(a0Selection())).Block(
					a0Name().Index(Id("i")).Op("=").Add(fn.CallerID()).Call(Id("each")),
				))
				continue
			}
			// Output:
			// aName, err := callerFn(a.Name)
			// if err != nil {
			//	return Bar{}, err
			// }
			g.Add(List(a0Name(), Id("err")).Op(":=").Add(fn.CallerID()).Call(a0Selection()))
			g.Add(genReturnOnError().Line())
		}
	}

	genReturnType := func(s *Statement) {
		if fn.Error != nil {
			// (Bar, nil)
			s.Add(Parens(List(genType(to.Type), Id("error"))))
		} else {
			// (Bar)
			s.Add(genType(to.Type))
		}
	}

	// No error signature for this function, however there are mappers with
	// errors.
	if fn.Error == nil && (len(methodsWithError)+len(mappersWithError) > 0) {
		panic(ErrMissingReturnError(fn))
	}

	// We need to know if the mapper has error signature.
	g.hasErrorByMapper[fn.NormalizedSignature()] = fn.Error != nil

	f.Func().
		Params(g.genShortName().Op("*").Id(typeName)). // (c *Converter)
		Id(fnName).                                    // mapMainAToMainB
		Params(
			Id(argsWithIndex(from.Name, 0)).Add(genType(from.Type)),
		). // (a A)
		Do(genReturnType).
		BlockFunc(func(g *Group) {
			// Output:
			// var a0Name *outpkg.OutType
			// if a0.Name != nil {
			//   res := methodcall(*a0.Name)
			//   a0Name = &res
			// }
			for _, pointer := range pointers {
				g.Add(pointer)
			}

			// Output:
			// a0Name, err := a0.Name()
			// if err != nil {
			//   return B{}, err
			// }
			for _, c := range methodsWithError {
				g.Add(c)
			}

			// Output:
			// aName, err := customfuncpkg.CustomFunc(a.Name)
			// aName, err := c.struct.CustomMethod(a.Name)
			// aName, err := c.mapAtoB(a.Name)
			// if err != nil {
			//	return Bar{}, err
			// }
			genPrivateMapperMethodsWithError(g)

			g.Add(ReturnFunc(func(g *Group) {
				if fn.Error != nil {
					// Output:
					// return Bar{}, nil
					g.Add(
						List(
							genType(to.Type).Values(dict),
							Id("nil"),
						),
					)
				} else {
					// Output:
					// return Bar{}
					g.Add(
						genType(to.Type).Values(dict))
				}
			}))
		}).Line()
	return f
}

func (g *Generator) genPublicMethod(f *jen.File, fn mapper.Func) {
	var (
		typeName = g.opt.TypeName
	)
	// Output:
	// func (c *Converter) Convert(a A) (B, error) {
	//   return c.mapMainAtoMainB(a.Field)
	// }

	this := g
	from, to := fn.From, fn.To
	if (from.Variadic || from.Type.IsSlice) != to.Type.IsSlice {
		panic("mapper: slice to no-slice and vice versa is not allowed")
	}
	isSlice := from.Type.IsSlice

	// main.A
	inType := genType(from.Type)

	// main.B
	outType := genType(to.Type)

	genInputType := func(g *Group) {
		g.Add(
			Id(argsWithIndex(from.Name, 0)).Do(func(s *Statement) {
				if !from.Variadic && from.Type.IsSlice {
					s.Add(Index())
				}
				if from.Type.IsPointer {
					s.Add(Op("*"))
				}
				if from.Variadic {
					s.Add(Op("..."))
				}
			}).Add(inType),
		)
	}

	genReturnType := func(s *Statement) {
		returnType := Do(func(rs *Statement) {
			// Output:
			// []main.B
			if to.Type.IsSlice {
				rs.Add(Index())
			}

			// Output:
			// []*main.B
			if to.Type.IsPointer {
				rs.Add(Op("*"))
			}
		}).Add(outType)

		if fn.Error != nil {
			// (*Bar, error)
			s.Add(Parens(List(returnType, Id("error"))))
		} else {
			// (*Bar)
			s.Add(returnType)
		}
	}

	genReturnOnError := func() *Statement {
		return If(Id("err").Op("!=").Id("nil")).Block(ReturnFunc(func(g *Group) {
			if isSlice || to.Type.IsPointer {
				g.Add(List(Id("nil"), Id("err")))
			} else {
				g.Add(List(outType.Clone().Values(), Id("err")))
			}
		})).Clone()
	}

	f.Func().
		Params(g.genShortName().Op("*").Id(typeName)). // (c *Converter)
		Id(fn.Name).ParamsFunc(genInputType).          // Convert(a *A)
		Do(genReturnType).                             // (*B, error)
		BlockFunc(func(g *Group) {
			genNameID := func() *Statement {
				return Id(argsWithIndex(from.Name, 0)).Clone()
			}
			mapperHasError := this.hasErrorByMapper[fn.NormalizedSignature()]
			if isSlice {
				// res := make([]B, len(a))
				// for i, each := range a {
				//   var err error
				//   res[i], err = c.mapMainAToMainB(each)
				//   if err != nil { return err }
				// }
				// return res, nil
				g.Add(Id("res").Op(":=").Make(List(Index().Add(outType), Len(genNameID()))))
				g.Add(For(List(Id("i"), Id("each")).Op(":=").Range().Add(genNameID())).BlockFunc(func(g *Group) {

					// If the private method does not have error, exit.
					if mapperHasError {
						g.Add(Var().Id("err").Id("error"))
						g.Add(List(
							Id("res").Index(Id("i")),
							Id("err"),
						).Op("=").Add(this.genShortName()).Dot(fn.NormalizedName()).Call(Id("each")))
						g.Add(genReturnOnError())
					} else {
						g.Add(Id("res").Index(Id("i")).Op("=").Add(this.genShortName()).Dot(fn.NormalizedName()).Call(Id("each")))
					}
				}))

				if fn.Error != nil {
					// return &res, nil
					g.Add(Return(List(Op(pointerOp(to.Type, "&")).Id("res"), Id("nil"))))
				} else {
					// return &res
					g.Add(Return(Op(pointerOp(to.Type, "&")).Id("res")))
				}
			} else {
				if to.Type.IsPointer {
					if fn.Error != nil {
						// Output:
						// res, err := c.mapMainAToMainB(a)
						// if err != nil {
						//   return nil, err
						// }
						// return c.mapMainAToMainB(a)
						g.Add(List(Id("res"), Id("err")).Op(":=").Add(this.genShortName()).Dot(fn.NormalizedName()).Call(genNameID()))
						g.Add(genReturnOnError())
						g.Add(Return(List(Op("&").Id("res")), Id("nil")))
					} else {
						// Output:
						// res := c.mapMainAToMainB(a)
						// return &res
						g.Add(Id("res").Op(":=").Add(this.genShortName()).Dot(fn.NormalizedName()).Call(genNameID()))
						g.Add(Return(Op("&").Id("res")))
					}
				} else {
					g.Add(Return(Op(pointerOp(to.Type, "&")).Add(this.genShortName()).Dot(fn.NormalizedName()).Call(genNameID())))
				}
			}
		}).Line()
}

func (g *Generator) genShortName() *Statement {
	return Id(mapper.ShortName(g.opt.TypeName)).Clone()
}

func pointerOp(m *mapper.Type, op string) string {
	if !m.IsPointer {
		return ""
	}
	return op
}

func argsWithIndex(name string, index int) string {
	return fmt.Sprintf("%s%d", name, index)
}

func (g *Generator) validateToAndFromStruct(fn mapper.Func) {
	from, to := fn.From, fn.To
	g.validateFieldMapping(from.Type, to.Type)

	fromFields := from.Type.StructFields
	fromMethods := from.Type.StructMethods

	// Check that the result struct has all the fields provided by the input
	// struct.
	for name, rhs := range to.Type.StructFields {
		if lhs, exists := fromFields[name]; exists {
			g.validateStructField(lhs, rhs)
			continue
		}
		if lhs, exists := fromMethods[name]; exists {
			g.validateMethodSignature(lhs, rhs)
			continue
		}
		panic(fmt.Sprintf("mapper: field not found: %s.%s does not have fields that maps to %s.%s(%s)",
			from.Type.Pkg,
			from.Name,
			to.Type.Pkg,
			name,
			to.Type.Type,
		))
	}
}

func (g *Generator) validateFieldMapping(lhs, rhs *mapper.Type) {
	if lhs.IsSlice != rhs.IsSlice {
		if lhs.IsSlice {
			// TODO: Add better message.
			panic("mapper: cannot convert from slice to struct")
		} else {
			panic("mapper: cannot convert from struct to slice")
		}
	}
	if lhs.IsPointer && !rhs.IsPointer {
		panic("mapper: value to pointer conversion not allowed")
	}
}

func (g *Generator) validateStructField(lhs, rhs mapper.StructField) {
	if !lhs.Type.Equal(rhs.Type) {
		// If one of the mappers already implement this, skip the error.
		if _, exists := g.mappers[buildFnSignature(lhs.Type, rhs.Type)]; exists {
			return
		}
		// There could also be a tag function.
		if lhs.Tag != nil {
			return
		}

		panic(ErrConversion(lhs.Type, rhs.Type))
	}
}

// validateMethodSignature checks if the lhs.method() returns the right
// signature required by rhs.
func (g *Generator) validateMethodSignature(lhs mapper.Func, rhs mapper.StructField) {
	if lhs.From != nil {
		panic(fmt.Sprintf("mapper: struct method should not have arguments %s.%s(%s %s)", lhs.Pkg, lhs.Name, lhs.From.Name, lhs.From.Type.Type))
	}

	// TODO: check if there is a local mapper that fulfils this type conversion.
	// This can only be from one of the converters.
	if !lhs.To.Type.Equal(rhs.Type) {
		panic(ErrConversion(lhs.To.Type, rhs.Type))
	}
}

func (g *Generator) validateFunctionSignatureMatch(fn *mapper.Func, lhs, rhs *mapper.Type) {
	var (
		in  = fn.From.Type
		out = fn.To.Type
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
}

func (g *Generator) validatePointerConversion(lhs, rhs *mapper.Type) {
	if lhs.IsPointer && !rhs.IsPointer {
		panic(fmt.Sprintf("mapper: conversion of value %s to pointer %s not allowed",
			lhs.Signature(),
			rhs.Signature(),
		))
	}
}

func (g *Generator) loadTagFunction(tag *mapper.Tag) *mapper.Func {
	fieldPkgPath := g.opt.PkgPath
	if tag.IsImported() {
		fieldPkgPath = tag.PkgPath
	}

	// Load the function.
	pkg := mapper.LoadPackage(fieldPkgPath)
	obj := mapper.LookupType(pkg, tag.Func)
	if obj == nil {
		panic(ErrFuncNotFound(tag))
	}

	fnType, ok := obj.(*types.Func)
	if !ok {
		panic(fmt.Sprintf("mapper: %q is not a func", tag.Func))
	}

	return mapper.ExtractFunc(fnType)
}

func buildFnSignature(lhs, rhs *mapper.Type) string {
	fn := mapper.Func{
		From: mapper.NewFuncArg("", lhs, false),
		To:   mapper.NewFuncArg("", rhs, false),
	}
	return fn.NormalizedSignature()
}

func buildType(t *mapper.Type) func(*Statement) {
	return func(s *Statement) {
		if t.IsSlice {
			s.Add(Index())
		}
		if t.IsPointer {
			s.Add(Op("*"))
		}
	}
}

func genType(T *mapper.Type) *Statement {
	if T.PkgPath != "" {
		return Qual(T.PkgPath, T.Type)
	}
	return Id(T.Type)
}

func ErrConversion(lhs, rhs *mapper.Type) error {
	return fmt.Errorf(`mapper: cannot convert %s to %s`,
		lhs.Signature(),
		rhs.Signature(),
	)
}

func ErrMismatchType(lhs, rhs *mapper.Type) error {
	return fmt.Errorf(`mapper: signature does not match: %s to %s`,
		lhs.Signature(),
		rhs.Signature(),
	)
}

func ErrMissingReturnError(fn mapper.Func) error {
	return fmt.Errorf("mapper: missing return err for %s", fn.PrettySignature())
}

func ErrFuncNotFound(tag *mapper.Tag) error {
	return fmt.Errorf("mapper: func %q from %s not found", tag.Func, tag.Tag)
}
