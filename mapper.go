package mapper

import (
	"flag"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

// StructField for the example below.
//type Foo struct {
//  Name sql.NullString `json:"name"
//}
type StructField struct {
	Name string `example:"Name"`
	// Useful when the output directory doesn't match the existing ones.
	PkgPath  string `example:"github.com/alextanhongpin/go-codegen/test"`
	PkgName  string `example:"test"`
	Exported bool   `example:"true"`
	Tag      string `example:"build:'-'"` // To ignore builder.
	*Field
	// StructMethod // Allow struct method name to be called too, e.g. Name() to match Name field.
}

func (f StructField) String() string {
	var result []string
	if f.PkgPath != "" {
		result = append(result, f.PkgPath)
	}
	if f.Name != "" {
		result = append(result, f.Name)
	}
	return strings.Join(result, ".")
}

type Field struct {
	Type         string `example:"NullString"`
	PkgPath      string `example:"database/sql"`
	IsPointer    bool
	IsCollection bool // Whether it's an array or slice.
	IsMap        bool
	MapKey       *Field
	MapValue     *Field
}

func (f Field) String() string {
	var result []string
	if f.PkgPath != "" {
		result = append(result, f.PkgPath)
	}
	if f.Type != "" {
		result = append(result, f.Type)
	}
	return strings.Join(result, ".")
}

// NewField recursively checks for the field type.
func NewField(typ types.Type) *Field {
	var isPointer, isCollection, isMap bool
	var fieldPkgPath, fieldType string
	var mapKey, mapValue *Field

	switch t := typ.(type) {
	case *types.Slice:
		isCollection = true
		typ = t.Elem()
	case *types.Array:
		isCollection = true
		typ = t.Elem()
	case *types.Map:
		isMap = true
		mapKey = NewField(t.Key())
		mapValue = NewField(t.Elem())
	}

	// In case the slice or array is pointer, we take the elem again.
	switch t := typ.(type) {
	case *types.Pointer:
		isPointer = true
		typ = t.Elem()
	}

	switch t := typ.(type) {
	case *types.Named:
		obj := t.Obj()
		fieldPkgPath = obj.Pkg().Path()
		fieldType = obj.Name()
	default:
		fieldType = t.String()
	}

	return &Field{
		Type:         fieldType,
		PkgPath:      fieldPkgPath,
		IsCollection: isCollection,
		IsPointer:    isPointer,
		IsMap:        isMap,
		MapKey:       mapKey,
		MapValue:     mapValue,
	}
}

type Option struct {
	In         string
	Out        string
	PkgName    string
	PkgPath    string
	StructName string
	Converters []ConverterFunc
}

type Struct struct {
	Name   string
	Field  *Field
	Fields map[string]StructField
}

type ConverterFunc struct {
	Name       string
	PkgPath    string
	From       *Struct
	To         *Struct
	HasError   bool
	HasContext bool
}

type Generator func(opt Option) error

func New(fn Generator) error {
	typePtr := flag.String("type", "", "the target type name")
	inPtr := flag.String("in", os.Getenv("GOFILE"), "the input file, defaults to the file with the go:generate comment")
	outPtr := flag.String("out", "", "the output directory")
	flag.Parse()

	in := fullPath(*inPtr)

	// Allows -type=Foo,Bar
	typeNames := strings.Split(*typePtr, ",")
	rootPkgPath := packagePath(in)

	for _, typeName := range typeNames {
		var out string
		if o := *outPtr; o == "" {
			// path/to/main.go becomes path/to/foo_gen.go
			out = filepath.Join(filepath.Dir(in), fileNameFromType(typeName))
		} else {
			if !hasExtension(o) {
				panic("mapper: out must be a valid go file")
			}
			out = fullPath(o)
		}

		pkg, inType := loadInterface(rootPkgPath, typeName) // github.com/your-github-username/your-pkg.
		//log.Printf("inPkg: %v\n", pkg)
		//log.Printf("inType: %v\n", inType)

		//fnPkg, fnType := loadFunction(rootPkgPath, "CustomConverter")
		//log.Printf("fnPkg: %v\n", fnPkg)
		//log.Printf("fnType: %v\n", fnType)

		converters := extractConverters(inType)
		if err := fn(Option{
			PkgName:    pkg.Name,
			PkgPath:    pkg.PkgPath,
			Out:        out,
			In:         in,
			StructName: typeName,
			Converters: converters,
		}); err != nil {
			return err
		}
	}
	return nil
}

func extractConverters(interfaceType *types.Interface) []ConverterFunc {
	n := interfaceType.NumExplicitMethods()

	//signature := make(map[string]string)
	var res []ConverterFunc

	for i := 0; i < n; i++ {
		field := interfaceType.ExplicitMethod(i)
		//fmt.Println("field.FullName:", field.FullName())
		//fmt.Println("field.Exported:", field.Exported()) // true
		//fmt.Println("field.Name:", field.Name())         // Convert.
		//fmt.Println("field.Pkg:", field.Pkg())
		//fmt.Println("field.Type:", field.Type().(*types.Signature))

		sigType := field.Type().(*types.Signature)
		//fmt.Println("sig.Params:", sigType.Params())
		//fmt.Println("sig.Params.Len:", sigType.Params().Len())

		param := sigType.Params().At(0)
		//debugVar("sig.Params.At(1)", param)

		//fmt.Println("sig.Results:", sigType.Results())
		//fmt.Println("sig.Results.Len:", sigType.Results().Len())
		result := sigType.Results().At(0)
		//debugVar("sig.Results.At(1)", result)
		//fmt.Println("sig.Variadic:", sigType.Variadic())

		from := NewField(param.Type())
		to := NewField(result.Type())
		//fmt.Println("from", from)
		//fmt.Println("to", to)

		// TODO: Handle slice to slice
		// TODO: Handle non-pointer to pointer

		// Check if it's a pointer first or not.
		paramType := param.Type().Underlying()
		if ptr, ok := paramType.(*types.Pointer); ok {
			fmt.Println("param is pointer")
			paramType = ptr.Elem()
		}
		paramStruct, ok := paramType.Underlying().(*types.Struct)
		if !ok {
			fmt.Println("param is not struct")
			continue
		}

		resultType := result.Type().Underlying()
		if ptr, ok := resultType.(*types.Pointer); ok {
			resultType = ptr.Elem()
		}
		resultStruct, ok := resultType.Underlying().(*types.Struct)
		if !ok {
			continue
		}

		paramFields := extractStructFields(paramStruct)
		resultFields := extractStructFields(resultStruct)
		//fmt.Println(paramFields, resultFields)

		if !compareStructFields(paramFields, resultFields) {
			panic(fmt.Sprintf("struct does not match %s", field.Name()))
		}

		convertFn := ConverterFunc{
			Name:    field.Name(),
			PkgPath: field.Pkg().Path(),
			From: &Struct{
				Name:   param.Name(),
				Field:  from,
				Fields: paramFields,
			},
			To: &Struct{
				Name:   result.Name(),
				Field:  to,
				Fields: resultFields,
			},
		}

		fmt.Println("convertFn", convertFn, convertFn.From, convertFn.To)
		res = append(res, convertFn)

		//signature[from.String()] = to.String()
		//fields[i] = StructField{
		//Name:     field.Name(),
		//PkgPath:  field.Pkg().Path(),
		//Exported: field.Exported(),
		//Field:    NewField(field.Type()),
		//}
		//fmt.Println("")
	}
	//fmt.Println(signature)
	return res
}

func debugVar(name string, v *types.Var) {
	fmt.Printf("%s: %v\n", name, v)
	fmt.Printf("%s.Anonymous(): %v\n", name, v.Anonymous())
	fmt.Printf("%s.IsField(): %v\n", name, v.IsField())
	fmt.Printf("%s.Exported(): %v\n", name, v.Exported())
	fmt.Printf("%s.Name(): %v\n", name, v.Name())
	fmt.Printf("%s.Pkg(): %v\n", name, v.Pkg())
	fmt.Printf("%s.Type(): %v\n", name, v.Type())
	fmt.Printf("%s.String(): %v\n", name, v.String())
	fmt.Printf("%s.NewField(): %+v\n", name, NewField(v.Type()))

	switch t := v.Type().Underlying().(type) {
	case *types.Pointer:
		fmt.Println("is pointer", t)
	case *types.Struct:
		fmt.Println("is struct", t)
	case *types.Slice:
		fmt.Println("is slice", t)
	case *types.Array:
		fmt.Println("is array", t)
	case *types.Map:
		fmt.Println("is map", t)
	default:
		fmt.Println("is unknown", t)
	}
}
