package main

import (
	"fmt"
	"go/types"
	"sort"

	"github.com/alextanhongpin/mapper"
	"github.com/alextanhongpin/mapper/cmd/mapper/internal"
	"github.com/alextanhongpin/mapper/loader"
	"github.com/dave/jennifer/jen"
	. "github.com/dave/jennifer/jen"
)

const GeneratorName = "github.com/alextanhongpin/mapper"

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if err := mapper.New(func(opt mapper.Option) error {
		gen := NewGenerator(opt)
		return gen.Generate()
	}); err != nil {
		fmt.Println(err)
	}
}

type Generator struct {
	opt              mapper.Option
	dependencies     map[string]types.Type
	mappers          map[string]bool
	hasErrorByMapper map[string]bool
	interfaceVisitor *internal.InterfaceVisitor
}

func NewGenerator(opt mapper.Option) *Generator {
	return &Generator{
		opt:              opt,
		dependencies:     make(map[string]types.Type),
		mappers:          make(map[string]bool),
		hasErrorByMapper: make(map[string]bool),
	}
}

func (g *Generator) Generate() error {
	var (
		pkgPath    = g.opt.PkgPath
		pkgName    = g.opt.PkgName
		interfaces = g.opt.Items
		out        = g.opt.Out
	)

	// If there is only one output, use the target struct
	// name. Otherwise, the filename will be used.
	if len(interfaces) == 1 {
		out = interfaces[0].Path
	}

	for _, opt := range interfaces {
		// Since a package path basename might not be the same as the package name,
		// This allows us to use Qual and exclude imports from the same package.
		f := NewFilePathName(pkgPath, pkgName)
		f.HeaderComment(fmt.Sprintf("Code generated by %s, DO NOT EDIT.", GeneratorName))

		iv := internal.NewInterfaceVisitor(opt.Type)
		interfaceMethods := iv.Methods()
		g.interfaceVisitor = iv

		// Cache first so that we can re-use later.
		var keys []string
		for key, method := range interfaceMethods {
			info, ok := iv.MethodInfo(method.Name)
			if !ok {
				panic(fmt.Errorf("method not found: %s", method.Name))
			}
			g.hasErrorByMapper[method.Normalize().Signature()] = info.HasError()
			keys = append(keys, key)
		}
		sort.Strings(keys)

		/*

			Collect the generated private methods, but not build them yet,
			mainly because there are some.
		*/
		var stmts []*Statement
		for _, key := range keys {
			method := interfaceMethods[key]
			signature := method.Normalize().Signature()
			if g.mappers[signature] {
				continue
			}
			stmt := g.genPrivateMethod(method, opt)
			stmts = append(stmts, stmt)
			g.mappers[signature] = true
		}

		// Generate the struct and constructor before the method declarations.
		g.genInterfaceChecker(f, opt)
		g.genStruct(f, opt)
		g.genConstructor(f, opt)

		for _, stmt := range stmts {
			f.Add(stmt)
		}

		for _, key := range keys {
			method := interfaceMethods[key]
			if !g.mappers[method.Normalize().Signature()] {
				panic("method not found")
			}
			g.genPublicMethod(f, method, opt)
		}

		if err := f.Save(out); err != nil { // e.g. main_gen.go
			return err
		}
	}
	fmt.Printf("success: generated %s\n", out)
	return nil
}

func (g *Generator) dependenciesKeys() []string {
	var keys []string
	for key := range g.dependencies {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (g *Generator) genInterfaceChecker(f *jen.File, opt mapper.OptionItem) {
	// Output:
	//
	// var _ Converter = (*ConverterImpl)(nil)

	f.Var().Op("_").Id(opt.Name).Op("=").Parens(Op("*").Id(g.genTypeName(opt))).Parens(Nil())
}

func (g *Generator) genStruct(f *jen.File, opt mapper.OptionItem) {
	// Output:
	//
	// type Converter struct {
	//   customInterface interfacepkg.CustomInterface
	//   customStruct    *structpkg.CustomStruct
	// }

	f.Type().Id(g.genTypeName(opt)).StructFunc(func(group *Group) {
		typeNames := g.dependenciesKeys()
		for _, typeName := range typeNames {
			use := g.dependencies[typeName]
			o := mapper.NewTypeName(use)
			p := o.Pkg()
			group.Add(Id(typeName), Do(func(s *Statement) {
				if mapper.IsStruct(use) {
					s.Add(Op("*"))
				}
			}).Qual(p.Path(), o.Name()))
		}
	}).Line()
}

func (g *Generator) genConstructor(f *jen.File, opt mapper.OptionItem) {
	// Output:
	//
	// func NewConverter(customStruct *structpkg.CustomStruct, customInterface interfacepkg.CustomInterface) *Converter {
	//   return &Converter{
	//     structpkgCustomStruct: customStruct,
	//     interfacepkgCustomInterface: customInterface,
	//   }
	// }

	typeName := g.genTypeName(opt)
	typeNames := g.dependenciesKeys()

	f.Func().Id(fmt.Sprintf("New%s", typeName)).ParamsFunc(func(group *Group) {
		for _, structName := range typeNames {
			use := g.dependencies[structName]
			o := mapper.NewTypeName(use)
			p := o.Pkg()
			group.Add(Id(structName), Do(func(s *Statement) {
				if mapper.IsStruct(use) {
					s.Add(Op("*"))
				}
			}).Qual(p.Path(), o.Name()))
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
func (g *Generator) genPrivateMethod(fn *mapper.Func, opt mapper.OptionItem) *jen.Statement {
	var (
		typeName      = g.genTypeName(opt)
		fnName        = fn.NormalizedName()
		from          = fn.From
		to            = fn.To
		methodInfo, _ = g.interfaceVisitor.MethodInfo(fn.Name)
		dict          = make(Dict)
	)

	// Since our private mapper does not have knowledge of error,
	// we need to set it manually.
	normFn := fn.Normalize()
	normFn.Error = methodInfo.HasError()

	// Loop through all the target keys.
	structFields := mapper.NewStructFields(to.Type).WithTags()
	keys := generateSortedStructFields(structFields)

	m := internal.NewMulti()
	for _, key := range keys {
		var r internal.Resolver
		// The RHS struct field.
		to := structFields[key]
		if to.Tag != nil && to.Tag.IsAlias() {
			key = to.Tag.Name
		}

		// If LHS field matches the RHS field ...
		if field, ok := methodInfo.Param.FieldByName(key); ok {
			// Just an ordinary LHS struct field. Noice.
			r = internal.NewFieldResolver(from.Name, field, to)
		}
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
		if method, ok := methodInfo.Param.MethodByName(key); ok {
			r = internal.NewMethodResolver(from.Name, method, to)
		}
		if r == nil {
			panic("not a field or method")
		}

		var (
			lhsType     types.Type
			tag         = r.Tag()
			rhsType     = r.Rhs().Type
			hasTag      = tag != nil
			bName       = func() *jen.Statement { return Id(r.Rhs().Name) }
			a0Name      = r.LhsVar
			a0Selection = r.RhsVar
		)

		funcBuilder := internal.NewFuncBuilder(r, normFn)

		if r.IsMethod() {
			// IS METHOD
			method := r.Lhs().(*mapper.Func)
			hasError := method.Error
			lhsType = method.To.Type

			// No tags, no errors, and equal types means we can assign the field directly.
			if !hasTag && !hasError && mapper.IsIdentical(lhsType, rhsType) {
				// Output:
				// Name: a0.Name()
				dict[bName()] = a0Selection()
				continue
			}

			if hasError {
				/*
					Output:

					a0Name, err := a0.Name()
					if err != nil {
						return B{}, err
					}
				*/
				m.Add(List(a0Name(), Err()).Op(":=").Add(a0Selection()))
				m.Add(funcBuilder.GenReturnOnError())
			} else {
				/*
					Output:

					a0Name := a0.Name()
				*/
				m.Add(a0Name().Op(":=").Add(a0Selection()))
			}

			// Don't exit yet, there might be another step of transformation.
			r.Assign()
		} else {
			// IS NOT METHOD A.K.A IS STRUCT FIELD
			lhs := r.Lhs().(mapper.StructField)
			lhsType = lhs.Type

			// No tags and equal types means we can assign the field directly.
			if !hasTag && mapper.IsUnderlyingIdentical(lhsType, rhsType) {
				if mapper.IsIdentical(lhsType, rhsType) {
					/*
						Output:

						B{
							Name: a0.Name(),
						}
					*/
					dict[bName()] = a0Selection()
				} else {
					// There may be non-pointer to pointer conversion, that wasn't
					// handled.
					if !mapper.IsPointer(lhsType) && mapper.IsPointer(rhsType) {
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

			// TAG: IS FUNC
			// The tag defines a custom function, TransformationFunc that can be used to
			// map LHS field to RHS.
			if tag.IsFunc() {
				fn, _ := methodInfo.Result.MapperByTag(tag.Tag)

				// Build the func.
				m.Add(funcBuilder.BuildFuncCall(fn, lhsType, rhsType))

				// The new type is the fn output type.
				lhsType = fn.To.Type
			}

			// TAG: IS METHOD
			// The tag loads a custom struct or interface method.
			if tag.IsMethod() {
				method, _ := methodInfo.Result.MapperByTag(tag.Tag)
				// To avoid different packages having same struct name, prefix the
				// struct name with the package name.
				g.dependencies[tag.Var()] = method.Obj.Type()

				m.Add(funcBuilder.BuildMethodCall(g.genShortName(opt).Dot(tag.Var()).Dot(method.Name), method, lhsType, rhsType))

				lhsType = method.To.Type
			}
		}

		if !mapper.IsUnderlyingIdentical(lhsType, rhsType) {
			// Check if there is a private mapper with the signature that accepts LHS
			// and returns RHS .
			signature := buildFnSignature(lhsType, rhsType)

			var method *mapper.Func
			interfaceMethods := mapper.NewInterfaceMethods(opt.Type)
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
			m.Add(funcBuilder.BuildMethodCall(g.genShortName(opt).Dot(method.Name), method, lhsType, rhsType))
			lhsType = method.To.Type
		}
		// RETURN VALUE.
		// bName: a0Name
		dict[bName()] = a0Selection()
	}

	return internal.NewMulti(
		Func().
			Params(g.genShortName(opt).Op("*").Id(typeName)).                      // (m *Converter)
			Id(fnName).                                                            // mapMainAToMainB
			Params(internal.GenInputType(internal.GenInputValue(normFn), normFn)). // (a A)
			Add(internal.GenReturnType(normFn)).
			BlockFunc(func(g *Group) {
				g.Add(m.Statement())

				returnType := internal.GenTypeName(to.Type).Values(dict)

				if normFn.Error {
					g.Add(Return(List(returnType, Nil())))
				} else {
					g.Add(Return(returnType))
				}
			})).Statement().
		Line()
}

func (g *Generator) genPublicMethod(f *jen.File, fn *mapper.Func, opt mapper.OptionItem) {
	var (
		typeName = g.genTypeName(opt)
		lhsType  = fn.From.Type
		rhsType  = fn.To.Type
	)

	lhs := mapper.StructField{
		Exported: true,
		Type:     mapper.NewUnderlyingType(lhsType),
	}
	rhs := mapper.StructField{
		Exported: true,
		Type:     mapper.NewUnderlyingType(rhsType),
	}

	res := internal.NewFieldResolver(fn.From.Name, lhs, rhs)
	arg := res.LhsVar()
	res.Assign()
	funcBuilder := internal.NewFuncBuilder(res, fn)

	normFn := fn.Normalize()
	normFn.Error = g.hasErrorByMapper[fn.Normalize().Signature()]
	method := funcBuilder.BuildMethodCall(g.genShortName(opt).Dot(normFn.NormalizedName()), normFn, lhsType, rhsType)

	f.Func().
		Params(g.genShortName(opt).Op("*").Id(typeName)). // (m *Converter)
		Id(fn.Name).
		Params(internal.GenInputType(arg, fn)). // Convert(a *A)
		Add(funcBuilder.GenReturnType()).       // (*B, error)
		BlockFunc(func(g *Group) {
			g.Add(method)

			if fn.Error {
				g.Add(Return(List(res.RhsVar(), Nil())))
			} else {
				g.Add(Return(res.RhsVar()))
			}
		}).Line()
}

func (g *Generator) genTypeName(opt mapper.OptionItem) string {
	return opt.Name + g.opt.Suffix
}

func (g *Generator) genShortName(opt mapper.OptionItem) *Statement {
	return Id(loader.ShortName(g.genTypeName(opt)))
}

func buildFnSignature(lhs, rhs types.Type) string {
	fn := mapper.NewFunc(mapper.NormFuncFromTypes("", lhs, rhs), nil)
	return fn.Normalize().Signature()
}

func generateSortedStructFields(fields mapper.StructFields) []string {
	var result []string
	for key := range fields {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
