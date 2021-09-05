package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/alextanhongpin/mapper"
	"github.com/alextanhongpin/mapper/cmd/mapper/internal"
	"github.com/dave/jennifer/jen"
	. "github.com/dave/jennifer/jen"
)

const GeneratorName = "github.com/alextanhongpin/mapper"

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
	interfaceVisitor *internal.InterfaceVisitor
}

func NewGenerator(opt mapper.Option) *Generator {
	return &Generator{
		opt:              opt,
		uses:             make(map[string]mapper.Type),
		mappers:          make(map[string]bool),
		hasErrorByMapper: make(map[string]bool),
	}
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

	iv := internal.NewInterfaceVisitor(typ.T)
	interfaceMethods := iv.Methods()
	g.interfaceVisitor = iv

	// Cache first so that we can re-use later.
	var keys []string
	for key, method := range interfaceMethods {
		signature := method.Normalize().Signature()
		g.mappers[signature] = false
		g.hasErrorByMapper[signature] = iv.MethodInfo()[method.Name].HasError()
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var stmts []*Statement
	for _, key := range keys {
		method := interfaceMethods[key]
		signature := method.Normalize().Signature()
		if g.mappers[signature] {
			continue
		}
		stmt := g.genPrivateMethod(method)
		stmts = append(stmts, stmt)
		g.mappers[signature] = true
	}

	// Generate the struct and constructor before the method declarations.
	g.genInterfaceChecker(f)
	g.genStruct(f)
	g.genConstructor(f)

	for _, stmt := range stmts {
		f.Add(stmt)
	}

	for _, key := range keys {
		method := interfaceMethods[key]
		if !g.mappers[method.Normalize().Signature()] {
			panic("mapper: method not found")
		}
		g.genPublicMethod(f, method)
	}

	if err := f.Save(out); err != nil { // e.g. main_gen.go
		return err
	}
	fmt.Printf("success: generated %s\n", out)
	return nil
}

func (g *Generator) usesKeys() []string {
	var keys []string
	for key := range g.uses {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (g *Generator) genInterfaceChecker(f *jen.File) {
	// Output:
	//
	// var _ Converter = (*ConverterImpl)(nil)

	f.Var().Op("_").Id(g.opt.TypeName).Op("=").Parens(Op("*").Id(g.genTypeName())).Parens(Nil())
}

func (g *Generator) genStruct(f *jen.File) {
	// Output:
	//
	// type Converter struct {
	//   customInterface interfacepkg.CustomInterface
	//   customStruct    *structpkg.CustomStruct
	// }

	f.Type().Id(g.genTypeName()).StructFunc(func(group *Group) {
		typeNames := g.usesKeys()
		for _, typeName := range typeNames {
			use := g.uses[typeName]
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
	//
	// func NewConverter(customStruct *structpkg.CustomStruct, customInterface interfacepkg.CustomInterface) *Converter {
	//   return &Converter{
	//     structpkgCustomStruct: customStruct,
	//     interfacepkgCustomInterface: customInterface,
	//   }
	// }

	typeName := g.genTypeName()
	typeNames := g.usesKeys()

	f.Func().Id(fmt.Sprintf("New%s", typeName)).ParamsFunc(func(group *Group) {
		for _, structName := range typeNames {
			use := g.uses[structName]
			group.Add(Id(structName), Do(func(s *Statement) {
				if use.IsStruct {
					s.Add(Op("*"))
				}
			}).Qual(use.PkgPath, use.Type))
		}
	}).Op("*").Id(typeName).Block(
		Return(Op("&").Id(typeName).ValuesFunc(func(group *Group) {
			dict := make(Dict)
			for _, structName := range typeNames {
				dict[Id(structName)] = Id(structName)
			}
			group.Add(dict)
		})),
	).Line()
}

// genPrivateMethod generates the most basic, struct A to struct B conversion
// without pointers, slice etc.
func (g *Generator) genPrivateMethod(fn *mapper.Func) *jen.Statement {
	var (
		f          = Null()
		typeName   = g.genTypeName()
		fnName     = fn.NormalizedName()
		from       = fn.From
		to         = fn.To
		methodInfo = g.interfaceVisitor.MethodInfo()[fn.Name]
	)

	// Loop through all the target keys.
	var keys []string
	for key := range to.Type.StructFields.WithTags() {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var resolvers []internal.Resolver
	for _, key := range keys {
		// The RHS struct field.
		to := to.Type.StructFields.WithTags()[key]

		// If LHS field matches the RHS field ...
		if field, ok := from.Type.StructFields[key]; ok {
			if to.Tag != nil && !to.Tag.IsField() {
				// Has a LHS struct field, but calls the method instead.
				//
				// Input:
				// type Lhs struct{
				//   // Custom `map` tag to indicate what method name to call.
				//   name string `map:"Name(),CustomFunc"`
				// }
				//
				// func (l Lhs) Name() string {}
				//
				structMethods := mapper.ExtractNamedMethods(from.Type.E)
				method := structMethods[key]
				resolvers = append(resolvers, internal.NewMethodResolver(from.Name, &field, method, to))
				continue
			}
			// Just an ordinary LHS struct field. Noice.
			resolvers = append(resolvers, internal.NewFieldResolver(from.Name, field, to))
		} else {
			// Has a LHS struct field, but calls the method instead.
			// The difference is there's no custom `map` tag to tell us what method it
			// is. Rather, we infer from the name of the RHS field.
			//
			// Input:
			// type Lhs struct{
			//   name string
			// }
			//
			// func (l Lhs) Name() string {}
			//
			// LHS method can also return error as the second argument.
			structMethods := mapper.ExtractNamedMethods(from.Type.E)
			if method, ok := structMethods[key]; ok {
				resolvers = append(resolvers, internal.NewMethodResolver(from.Name, nil, method, to))
			}
		}
	}

	var c internal.C
	dict := make(Dict)
	for _, r := range resolvers {
		var (
			lhsType     *mapper.Type
			tag         = r.Tag()
			rhsType     = r.Rhs().Type
			hasTag      = tag != nil
			bName       = func() *jen.Statement { return Id(r.Rhs().Name) }
			a0Name      = r.LhsVar
			a0Selection = r.RhsVar
		)

		funcBuilder := internal.NewFuncBuilder(r, fn)

		if r.IsMethod() {
			// IS METHOD
			method := r.Lhs().(*mapper.Func)
			hasError := method.Error
			lhsType = method.To.Type

			// No tags, no errors, and equal types means we can assign the field directly.
			if !hasTag && !hasError && lhsType.Equal(rhsType) {
				// Output:
				// Name: a0Name.Name()
				dict[bName()] = a0Selection()
				continue
			}

			if hasError {
				// Output:
				//
				// a0Name, err := a0Name.Name()
				// if err != nil { ...  }
				c.Add(List(a0Name(), Err()).Op(":=").Add(a0Selection()))
				c.Add(funcBuilder.GenReturnOnError())
			} else {
				// Output:
				// a0Name := a0Name.Name()
				c.Add(a0Name().Op(":=").Add(a0Selection()))
			}
			// Don't exit yet, there might be another step of transformation.
			r.Assign()
		} else {
			// IS NOT METHOD | IS STRUCT FIELD
			lhs := r.Lhs().(mapper.StructField)
			lhsType = lhs.Type

			// No tags and equal types means we can assign the field directly.
			if !hasTag && lhsType.EqualElem(rhsType) {
				if lhsType.Equal(rhsType) {
					// Output:
					//
					// Name: a0Name.Name
					dict[bName()] = a0Selection()
				} else {
					if !lhsType.IsPointer && rhsType.IsPointer {
						dict[bName()] = Op("&").Add(a0Selection())
					} else {
						dict[bName()] = a0Selection()
					}
				}
				continue
			}
			// There are probably further conversion for this field.
		}

		// METHOD OR FIELD RESOLVED.

		// TAG.
		// A tag exists, and could have transformation functions.
		if tag != nil && tag.HasFunc() {
			// If a method is provided, it works for single or slice, but the output
			// raw type must match.

			// TAG: FUNC
			// The tag defines a custom function, TransformationFunc that can be used to
			// map LHS field to RHS.
			if tag.IsFunc() {
				fn, _ := methodInfo.Result.MapperByTag(tag.Tag)

				// Build the func.
				funcBuilder.BuildFuncCall(&c, fn, lhsType, rhsType)

				// The new type is the fn output type.
				lhsType = fn.To.Type
			}

			// TAG: IS METHOD
			// The tag loads a custom struct or interface method.
			if tag.IsMethod() {
				method, _ := methodInfo.Result.MapperByTag(tag.Tag)
				typ := mapper.NewType(method.Obj.Type())
				// To avoid different packages having same struct name, prefix the
				// struct name with the package name.
				g.uses[tag.Var()] = *typ

				funcBuilder.BuildMethodCall(&c, g.genShortName().Dot(tag.Var()).Dot(method.Name), method, lhsType, rhsType)
				lhsType = method.To.Type
			}
		}

		if !lhsType.EqualElem(rhsType) {
			// Check if there is a private mapper with the signature that accepts LHS
			// and returns RHS .
			signature := buildFnSignature(lhsType, rhsType)
			if _, ok := g.mappers[signature]; !ok {
				panic(ErrConversion(lhsType, rhsType))
			}

			var method *mapper.Func
			interfaceMethods := internal.GenerateInterfaceMethods(g.opt.Type.T)
			for _, met := range interfaceMethods {
				if met.Normalize().Signature() == signature {
					method = met.Normalize()
					// Private mapper does not have error signature.
					// Therefor, we have to manually assign them.
					method.Error = g.hasErrorByMapper[signature]
					break
				}
			}
			// Method found.
			funcBuilder.BuildMethodCall(&c, g.genShortName().Dot(method.Name), method, lhsType, rhsType)
			lhsType = method.To.Type
		}
		// RETURN VALUE.
		// bName: a0Name
		dict[bName()] = a0Selection()
	}

	// Since our private mapper does not have knowledge of error,
	// we need to set it manually.
	//g.hasErrorByMapper[fn.Normalize().Signature()] = hasError

	normFn := fn.Normalize()
	normFn.Error = methodInfo.HasError()

	f.Func().
		Params(g.genShortName().Op("*").Id(typeName)).                         // (c *Converter)
		Id(fnName).                                                            // mapMainAToMainB
		Params(internal.GenInputType(internal.GenInputValue(normFn), normFn)). // (a A)
		Add(internal.GenReturnType(normFn)).
		BlockFunc(func(g *Group) {
			for _, code := range c {
				g.Add(code)
			}

			g.Add(Return(
				List(
					internal.GenTypeName(to.Type).Values(dict),
					Do(func(s *Statement) {
						if normFn.Error {
							s.Add(Nil())
						}
					}),
				),
			))
		}).Line()
	return f
}

func (g *Generator) genPublicMethod(f *jen.File, fn *mapper.Func) {
	var (
		typeName = g.genTypeName()
		lhsType  = fn.From.Type
		rhsType  = fn.To.Type
	)

	// TODO: REPLACE WITH FUNC
	this := g

	lhs := mapper.StructField{
		Name:     lhsType.Type,
		Pkg:      lhsType.Pkg,
		PkgPath:  lhsType.PkgPath,
		Exported: true,
		Tag:      nil,
		Type:     lhsType.Normalize(),
	}
	rhs := mapper.StructField{
		Name:     rhsType.Type,
		Pkg:      rhsType.Pkg,
		PkgPath:  rhsType.PkgPath,
		Exported: true,
		Tag:      nil,
		Type:     rhsType.Normalize(),
	}

	var c internal.C
	res := internal.NewFieldResolver(fn.From.Name, lhs, rhs)
	arg := res.LhsVar()
	res.Assign()
	funcBuilder := internal.NewFuncBuilder(res, fn)

	mapperHasError := g.hasErrorByMapper[fn.Normalize().Signature()]
	if mapperHasError && !fn.Error {
		panic(ErrMissingReturnError(fn))
	}

	f.Func().
		Params(g.genShortName().Op("*").Id(typeName)). // (c *Converter)
		Id(fn.Name).
		Params(internal.GenInputType(arg, fn)). // Convert(a *A)
		Add(funcBuilder.GenReturnType()).       // (*B, error)
		BlockFunc(func(g *Group) {
			normFn := fn.Normalize()
			normFn.Error = mapperHasError

			funcBuilder.BuildMethodCall(&c, this.genShortName().Dot(normFn.NormalizedName()), normFn, lhsType, rhsType)

			for _, code := range c {
				g.Add(code)
			}

			if fn.Error {
				g.Add(Return(List(res.RhsVar(), Nil())))
			} else {
				g.Add(Return(res.RhsVar()))
			}
		}).Line()
}

func (g *Generator) genTypeName() string {
	return g.opt.TypeName + g.opt.Suffix
}

func (g *Generator) genShortName() *Statement {
	return Id(mapper.ShortName(g.genTypeName()))
}

func pointerOp(m *mapper.Type, op string) string {
	if !m.IsPointer {
		return ""
	}
	return op
}

func buildFnSignature(lhs, rhs *mapper.Type) string {
	fn := mapper.NewFunc(mapper.NormFuncFromTypes(lhs, rhs))
	return fn.Normalize().Signature()
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

func ErrMissingReturnError(fn *mapper.Func) error {
	return fmt.Errorf("mapper: missing return err for %s", fn.Signature())
}

func ErrFuncNotFound(tag *mapper.Tag) error {
	return fmt.Errorf("mapper: func %q from %s not found", tag.Func, tag.Tag)
}
