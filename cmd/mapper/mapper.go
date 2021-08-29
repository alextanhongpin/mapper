package main

import (
	"fmt"
	"go/types"
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

	// Cache first so that we can re-use later.
	var keys []string
	for key, method := range typ.InterfaceMethods {
		g.mappers[method.Normalize().Signature()] = false
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var stmts []*Statement
	for _, key := range keys {
		method := typ.InterfaceMethods[key]
		g.validateToAndFromStruct(method)
		signature := method.Normalize().Signature()
		if g.mappers[signature] {
			continue
		}
		stmt := g.genPrivateMethod(method)
		stmts = append(stmts, stmt)
		g.mappers[signature] = true
	}

	// Generate the struct and constructor before the method declarations.
	g.genStruct(f)
	g.genConstructor(f)

	for _, stmt := range stmts {
		f.Add(stmt)
	}

	for _, key := range keys {
		method := typ.InterfaceMethods[key]
		if !g.mappers[method.Normalize().Signature()] {
			panic("mapper: method not found")
		}
		g.genPublicMethod(f, method)
	}

	return f.Save(out) // e.g. main_gen.go
}

func (g *Generator) usesKeys() []string {
	var keys []string
	for key := range g.uses {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (g *Generator) genStruct(f *jen.File) {
	// Output:
	//
	// type Converter struct {
	//   customInterface interfacepkg.CustomInterface
	//   customStruct    *structpkg.CustomStruct
	// }

	f.Type().Id(g.opt.TypeName).StructFunc(func(group *Group) {
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

	typeName := g.opt.TypeName
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
func (g *Generator) genPrivateMethod(fn mapper.Func) *jen.Statement {
	var (
		f              = Null()
		typeName       = g.opt.TypeName
		fnName         = fn.NormalizedName()
		from           = fn.From
		to             = fn.To
		parentHasError = fn.Error
	)

	// Loop through all the target keys.
	var keys []string
	for key := range to.Type.StructFields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var resolvers []internal.Resolver
	for _, key := range keys {
		// The RHS struct field.
		to := to.Type.StructFields[key]

		// If LHS field matches the RHS field ...
		if field, ok := from.Type.StructFields[key]; ok {
			if field.Tag != nil && !field.Tag.IsField() {
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
				method, ok := structMethods[key]
				if !ok {
					panic(fmt.Sprintf("mapper: method not found: %s", field.Tag.Name))
				}

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
			} else {
				panic(fmt.Sprintf("mapper: cannot map field %q for %s", key, to.Type.Signature()))
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

		funcBuilder := internal.NewFuncBuilder(r, &fn)

		if r.IsMethod() {
			// IS METHOD
			method := r.Lhs().(mapper.Func)
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
				if !parentHasError {
					panic(ErrMissingReturnError(fn))
				}
				// Output:
				//
				// a0Name, err := a0Name.Name()
				// if err != nil { ...  }
				c.Add(List(a0Name(), Id("err")).Op(":=").Add(a0Selection()))
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
				lhs := r.Lhs().(mapper.StructField)
				fn := g.loadTagFunction(&lhs)

				if fn.Error {
					// TODO: FIXME
					if !parentHasError {
						panic(ErrMissingReturnError(*fn))
					}
				}

				// Build the func.
				funcBuilder.BuildFuncCall(&c, fn, lhsType, rhsType)

				// The new type is the fn output type.
				lhsType = fn.To.Type
			}

			// TAG: IS METHOD
			// The tag loads a custom struct or interface method.
			if tag.IsMethod() {
				fieldPkgPath := lhsType.PkgPath
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
					structMethods := mapper.ExtractNamedMethods(typ.E)
					method = structMethods[tag.Func]
				default:
					panic(fmt.Sprintf("mapper: tag %q is invalid", tag.Tag))
				}
				g.validateFunctionSignatureMatch(&method, lhsType, rhsType)
				if method.Error && !parentHasError {
					panic(ErrMissingReturnError(fn))
				}

				// To avoid different packages having same struct name, prefix the
				// struct name with the package name.
				g.uses[tag.Var()] = *typ

				funcBuilder.BuildMethodCall(&c, g.genShortName().Dot(tag.Var()).Dot(method.Name), &method, lhsType, rhsType)
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
			for _, met := range g.opt.Type.InterfaceMethods {
				if met.Normalize().Signature() == signature {
					method = met.Normalize()
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

	// No error signature for this function, however there are mappers with
	// errors.
	// TODO:

	// We need to know if the mapper has error signature.
	g.hasErrorByMapper[fn.Normalize().Signature()] = fn.Error
	returnType := func() *Statement { return internal.GenTypeName(to.Type).Clone() }

	f.Func().
		Params(g.genShortName().Op("*").Id(typeName)).                                // (c *Converter)
		Id(fnName).                                                                   // mapMainAToMainB
		Params(Id(argsWithIndex(from.Name, 0)).Add(internal.GenTypeName(from.Type))). // (a A)
		Do(func(s *Statement) {
			// Return type must not be pointer.
			if fn.Error {
				// Output:
				// (B, error)
				s.Add(Parens(List(returnType(), Id("error"))))
			} else {
				// Output:
				// B
				s.Add(returnType())
			}
		}).
		BlockFunc(func(g *Group) {
			for _, code := range c {
				g.Add(code)
			}
			g.Add(ReturnFunc(func(g *Group) {
				if fn.Error {
					// Output:
					// return Bar{}, nil
					g.Add(List(returnType().Values(dict), Id("nil")))
				} else {
					// Output:
					// return Bar{}
					g.Add(returnType().Values(dict))
				}
			}))
		}).Line()
	return f
}

func (g *Generator) genPublicMethod(f *jen.File, fn mapper.Func) {
	var (
		typeName = g.opt.TypeName
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
	funcBuilder := internal.NewFuncBuilder(res, &fn)

	// TODO: Separate validation.
	//if many != rhs.IsSlice {
	//panic("mapper: slice to no-slice and vice versa is not allowed")
	//}

	mapperHasError := g.hasErrorByMapper[fn.Normalize().Signature()]
	if mapperHasError && !fn.Error {
		panic(fmt.Sprintf("mapper: missing return error for %s", fn.Signature()))
	}

	f.Func().
		Params(g.genShortName().Op("*").Id(typeName)).               // (c *Converter)
		Id(fn.Name).Params(arg.Add(internal.GenType(fn.From.Type))). // Convert(a *A)
		Add(funcBuilder.GenReturnType()).                            // (*B, error)
		BlockFunc(func(g *Group) {
			normFn := fn.Normalize()
			normFn.Error = mapperHasError

			funcBuilder.BuildMethodCall(&c, this.genShortName().Dot(normFn.NormalizedName()), normFn, lhsType, rhsType)

			for _, code := range c {
				g.Add(code)
			}

			if fn.Error {
				g.Add(Return(List(res.RhsVar(), Id("nil"))))
			} else {
				g.Add(Return(res.RhsVar()))
			}
		}).Line()
}

func (g *Generator) genShortName() *Statement {
	return Id(mapper.ShortName(g.opt.TypeName))
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
	fromMethods := mapper.ExtractNamedMethods(from.Type.E)

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
		// There could also be a value to pointer conversion.
		// Only applies for the same type, e.g. string to *string.
		// For structs, they may belong to different package.
		if (lhs.Type.Type == rhs.Type.Type) && (lhs.Type.PkgPath == rhs.Type.PkgPath) && (!lhs.IsPointer && rhs.IsPointer) {
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

// validateFunctionSignatureMatch ensures that the conversion from input to
// output is allowed.
// Function can receive value/pointer.
// Function must return a value/pointer with optional error.
// Function must accept the input signature of the type.
// Input can be one or many.
// Function must accept struct only, if the input is many, it will be
// operated at elem level.
func (g *Generator) validateFunctionSignatureMatch(fn *mapper.Func, lhs, rhs *mapper.Type) {
	var (
		in                  = fn.From.Type
		out                 = fn.To.Type
		pointerToNonPointer = lhs.IsPointer && !rhs.IsPointer
		isMany              = fn.From.Type.IsSlice || fn.From.Variadic
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

	if isMany {
		panic(fmt.Sprintf("mapper: func input must be struct: %s, provided %s", fn.Signature(), lhs.Signature()))
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

func (g *Generator) loadTagFunction(field *mapper.StructField) *mapper.Func {
	tag := field.Tag
	// Use the field pkg path from where the left function
	// reside. It may be on different files.
	fieldPkgPath := field.PkgPath
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

	return mapper.NewFunc(fnType)
}

func buildFnSignature(lhs, rhs *mapper.Type) string {
	fn := mapper.NewFunc(mapper.NormFuncFromTypes(lhs, rhs))
	return fn.Normalize().Signature()
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
	return fmt.Errorf("mapper: missing return err for %s", fn.Signature())
}

func ErrFuncNotFound(tag *mapper.Tag) error {
	return fmt.Errorf("mapper: func %q from %s not found", tag.Func, tag.Tag)
}
