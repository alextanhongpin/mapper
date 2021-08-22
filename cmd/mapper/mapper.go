package main

import (
	"fmt"
	"go/types"
	"log"

	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
	. "github.com/dave/jennifer/jen"
)

const GeneratorName = "mapper"

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
		pkgName = g.opt.PkgName
		typ     = g.opt.Type
		out     = g.opt.Out
	)

	f := NewFile(pkgName) // e.g. main
	f.PackageComment(fmt.Sprintf("Code gend by %s, DO NOT EDIT.", GeneratorName))

	// Cache first so that we can re-use later.
	for _, method := range typ.InterfaceMethods {
		g.mappers[method.NormalizedSignature()] = false
	}

	var stmts []*Statement
	for _, method := range typ.InterfaceMethods {
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

	for _, method := range typ.InterfaceMethods {
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
			}).Qual(relativeTo(g.opt.PkgPath, use.PkgPath), use.Type))
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
			}).Qual(relativeTo(g.opt.PkgPath, use.PkgPath), use.Type))
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

// genPrivateMethod gens the most basic type for the given function to be
// reused for slice mapping.
func (g *Generator) genPrivateMethod(f *jen.Statement, fn mapper.Func) *jen.Statement {
	// Output:
	// func (c *Converter) mapMainAToMainB(a A) (B, error) {
	//   aID, err := CustomFunc(a.ID)
	//   if err != nil {
	//     return B{}, err
	//   }
	//   return B{
	//     ID: a.ID,
	//     Name: a.Name,
	//   }, nil
	// }

	var (
		pkgPath  = g.opt.PkgPath
		typeName = g.opt.TypeName
		fnName   = fn.NormalizedName()
		from     = fn.From
		to       = fn.To
		this     = g
	)

	var methodsWithError []FieldResolver
	var mappersWithError []mapperFunc
	var resolvers []FieldResolver

	for key, to := range to.Type.StructFields {
		// Check if there is a field mapping.
		field, ok := from.Type.StructFields[key]
		if ok {
			resolver := NewStructFieldResolver(&field, from.Name, to)
			resolvers = append(resolvers, resolver)
			continue
		}

		// Check if there is a method with the name that returns the same
		// signature.
		// The name of the method matches the name of field, e.g. lhs.Age:
		// rhs.Age()
		if method, ok := from.Type.StructMethods[key]; ok {
			resolver := NewMethodFieldResolver(&method, from.Name, to)
			resolvers = append(resolvers, resolver)
			continue
		}
		panic("mapper: method signature not found")
	}

	dict := make(Dict)
	for _, resolver := range resolvers {
		rhs := resolver.Rhs()

		if resolver.IsMethod() {
			method := resolver.Lhs().(*mapper.Func)
			if method.Error != nil {
				// name: a0Name
				dict[Id(rhs.Name)] = resolver.Var()
				methodsWithError = append(methodsWithError, resolver)
			} else {
				// name: a0.Name()
				dict[Id(rhs.Name)] = resolver.Selection()
			}
			continue
		}

		lhs := resolver.Lhs().(*mapper.StructField)

		// `map:"CustomField,CustomFunc"`
		if tag := lhs.Tag; tag != nil && tag.HasFunc() {
			if tag.IsFunc() {
				fn := g.loadTagFunction(tag)
				g.validateFunctionSignatureMatch(fn, lhs.Type, rhs.Type)

				if fn.Error != nil {
					mappersWithError = append(mappersWithError, mapperFunc{
						Fn:       fn,
						In:       *lhs,
						resolver: resolver,
						// funcpkg.CustomFunc
						callerID: Qual(fn.PkgPath, fn.Name),
					})
					// Name: aName,
					dict[Id(rhs.Name)] = resolver.Var()
					continue
				}

				// name: customfuncpkg.CustomFunc(a.Name)
				dict[Id(rhs.Name)] = Qual(relativeTo(g.opt.PkgPath, fn.PkgPath), fn.Name).Call(resolver.Selection())
			}

			if tag.IsMethod() {
				// Could either be a struct or interface.
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
					panic("mapper: not implemented")
				}

				g.validateFunctionSignatureMatch(&method, lhs.Type, rhs.Type)

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
						In:         *lhs,
						structName: structPackageName,
						resolver:   resolver,
						// c.structpkgCustomStruct.CustomMethod(a0.Name)
						callerID: g.genShortName().Dot(structPackageName).Dot(method.Name).Call(resolver.Selection()),
					})
					// Name: aName,
					dict[Id(rhs.Name)] = resolver.Var()
					continue
				}

				// Name: c.interfacePkgInterface.CustomMethod(a.Name)
				dict[Id(rhs.Name)] = g.genShortName().Dot(structPackageName).Dot(method.Name).Call(resolver.Selection())
			}
			continue
		} // End of custom tag function.

		if !lhs.Type.Equal(rhs.Type) {
			methodSignature := buildFnSignature(lhs.Type, rhs.Type)
			_, hasSignature := g.mappers[methodSignature]
			if !hasSignature {
				panic(ErrConversion(lhs.Type, rhs.Type))
			}

			// Implement conversion for that field using existing converters.
			if method, exists := g.opt.Type.InterfaceMethods[methodSignature]; exists {
				if method.Error != nil || rhs.Type.IsSlice {
					// Name: a0Name,
					dict[Id(rhs.Name)] = resolver.Var()
					mappersWithError = append(mappersWithError, mapperFunc{
						Fn:       &method,
						In:       *lhs,
						resolver: resolver,
						// c.mapAToB
						callerID: g.genShortName().Dot(method.NormalizedName()),
					})
				} else {
					// Name: c.mapAtoB(a.Name)
					dict[Id(rhs.Name)] = g.genShortName().Dot(method.NormalizedName()).Call(resolver.Selection())
				}
				break
			}
			continue
		}

		// Output:
		// Name: a.Name
		dict[Id(rhs.Name)] = resolver.Selection()
		continue
	}

	genReturn := func() *Statement {
		return If(Id("err").Op("!=").Id("nil").Block(
			Return(
				List(
					Op(pointerOp(to.Type, "&")).Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(Dict{}),
					Id("err"),
				),
			)),
		).Clone()
	}

	// For each methods that has error as the second return value,
	// initialize them and return it.
	genMethodsWithError := func(g *Group) {
		for _, res := range methodsWithError {
			// a0Name, err := a0.Name()
			// if err != nil {
			//	return Bar{}, err
			// }
			g.Add(List(res.Var(), Id("err")).Op(":=").Add(res.Selection()))
			g.Add(genReturn().Line())
		}
	}

	genPrivateMapperMethodsWithError := func(g *Group) {
		for _, fn := range mappersWithError {
			// []*B
			returnType := Do(func(s *Statement) {
				if fn.In.Type.IsSlice {
					s.Add(Index())
				}
				if fn.In.Type.IsPointer {
					s.Add(Op("*"))
				}
			}).Qual(relativeTo(this.opt.PkgPath, fn.Fn.To.Type.PkgPath), fn.Fn.To.Type.Type)

			if fn.In.Type.IsSlice {
				// aName := make([]B, len(a.Name))
				g.Add(fn.resolver.Var().Op(":=").Make(returnType, Len(fn.resolver.Selection())))

				if fn.Fn.Error != nil {
					// Output:
					// aName := make([]B, len(a.Name))
					// for i, each := range a.Name {
					//   var err error
					//   aName[i], err = callerFunc(a.Name)
					//   if err != nil {
					//     return Bar{}, err
					//   }
					// }

					g.Add(For(List(Id("i"), Id("each")).Op(":=").Range().Add(fn.resolver.Selection())).Block(
						Var().Id("err").Id("error"),
						List(fn.resolver.Var().Index(Id("i")), Id("err")).Op("=").Add(fn.CallerID()).Call(Id("each")),
						genReturn(),
					).Line())
				} else {
					// Output:
					// aName := make([]B, len(a.Name))
					// for i, each := range a.Name {
					//   aName[i] := callerFn(a.Name)
					// }
					g.Add(For(List(Id("i"), Id("each")).Op(":=").Range().Add(fn.resolver.Selection())).Block(
						fn.resolver.Var().Index(Id("i")).Op("=").Add(fn.CallerID()).Call(Id("each")),
					))
				}
			} else {
				// Output:
				// aName, err := callerFn(a.Name)
				// if err != nil {
				//	return Bar{}, err
				// }
				g.Add(List(fn.resolver.Var(), Id("err")).Op(":=").Add(fn.CallerID()).Call(fn.resolver.Selection()))
				g.Add(genReturn().Line())
			}
		}
	}

	genReturnType := func(s *Statement) {
		returnType := Do(func(s *Statement) {
			if to.Type.IsPointer {
				s.Add(Op("*"))
			}
		}).Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type)
		if fn.Error != nil {
			// (Bar, nil)
			s.Add(Parens(List(
				returnType,
				Id("error"),
			)))
		} else {
			// (Bar)
			s.Add(returnType)
		}
	}

	// No error signature for this function, however there are mappers with
	// errors.
	if fn.Error == nil && (len(methodsWithError)+len(mappersWithError) > 0) {
		panic(ErrMissingReturnError(fn))
	}
	g.hasErrorByMapper[fn.NormalizedSignature()] = fn.Error != nil

	f.Func().
		Params(g.genShortName().Op("*").Id(typeName)). // (c *Converter)
		Id(fnName).                                    // mapMainAToMainB
		Params(
			Id(argsWithIndex(from.Name, 0)).Qual(relativeTo(pkgPath, from.Type.PkgPath), from.Type.Type),
		). // (a A)
		Do(genReturnType).
		BlockFunc(func(g *Group) {
			// Output:
			// aName, err := a.Name()
			// if err != nil {
			//	return Bar{}, err
			// }
			genMethodsWithError(g)

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
					// return Bar{}, nil
					g.Add(
						List(
							Op(pointerOp(to.Type, "&")).
								Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(dict),
							Id("nil"),
						),
					)
				} else {
					// return Bar{}
					g.Add(
						Op(pointerOp(to.Type, "&")).
							Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type).Values(dict))
				}
			}))
		}).Line()
	return f
}

func (g *Generator) genPublicMethod(f *jen.File, fn mapper.Func) {
	var (
		pkgPath  = g.opt.PkgPath
		typeName = g.opt.TypeName
	)
	// Output:
	// func (c *Converter) Convert(a A) (B, error) {
	//   b, err := c.mapMainAtoMainB(a.Field)
	//   return b, err
	// }

	this := g
	from, to := fn.From, fn.To
	if (from.Variadic || from.Type.IsSlice) != to.Type.IsSlice {
		panic("mapper: slice to no-slice and vice versa is not allowed")
	}
	isSlice := from.Type.IsSlice

	// main.A
	inType := Qual(relativeTo(pkgPath, from.Type.PkgPath), from.Type.Type)

	// main.B
	outType := Qual(relativeTo(pkgPath, to.Type.PkgPath), to.Type.Type)

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
		}).Op(pointerOp(to.Type, "*")).Add(outType)

		if fn.Error != nil {
			// (*Bar, error)
			s.Add(Parens(List(returnType, Id("error"))))
		} else {
			// (*Bar)
			s.Add(returnType)
		}
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
						g.Add(If(Id("err").Op("!=").Id("nil")).Block(ReturnFunc(func(g *Group) {
							if isSlice {
								g.Add(List(Id("nil"), Id("err")))
							} else {
								g.Add(List(outType.Clone().Values(), Id("err")))
							}
						})))
					} else {
						g.Add(Id("res").Index(Id("i")).Op("=").Add(this.genShortName()).Dot(fn.NormalizedName()).Call(Id("each")))
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
				// return c.mapMainAToMainB(a)
				g.Add(Return(this.genShortName().Dot(fn.NormalizedName()).Call(genNameID())))
			}
		}).Line()
}

func (g *Generator) genShortName() *Statement {
	return Id(mapper.ShortName(g.opt.TypeName)).Clone()
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
	if lhs.IsPointer || rhs.IsPointer {
		panic("mapper: struct pointer is not allowed")
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

func (g *Generator) validateFunctionSignatureMatch(fn *mapper.Func, in, out *mapper.Type) {
	if !fn.From.Type.Equal(in) {
		if fn.From.Type.Type != in.Type {
			panic(ErrMismatchType(fn.From.Type, in))
		}
	}

	if !fn.To.Type.Equal(out) {
		if fn.To.Type.Type != out.Type {
			panic(ErrMismatchType(fn.To.Type, out))
		}
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
