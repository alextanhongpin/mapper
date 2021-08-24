package mapper

import (
	"go/types"
	"strings"
)

type Type struct {
	Type             string `example:"NullString"`
	Pkg              string `example:"sql"`
	PkgPath          string `example:"database/sql"`
	IsStruct         bool
	IsPointer        bool
	IsArray          bool // Whether it's an array or slice.
	IsInterface      bool
	IsSlice          bool
	IsError          bool
	IsMap            bool
	MapKey           *Type
	MapValue         *Type
	StructFields     map[string]StructField
	InterfaceMethods map[string]Func
	T                types.Type
}

// Signature is used to compare if two types are equal.
func (t Type) Signature() string {
	// *github.com/alextanhongpin/yourpkg/Bar
	var parts []string
	if t.IsPointer {
		parts = append(parts, "*")
	}
	if t.IsSlice {
		parts = append(parts, "[]")
	}
	if t.PkgPath != "" {
		parts = append(parts, t.PkgPath)
	}
	parts = append(parts, t.Type)
	return strings.Join(parts, "")
}

func (t Type) Equal(other *Type) bool {
	return t.Signature() == other.Signature()
}

// NewType recursively checks for the field type.
func NewType(typ types.Type) *Type {
	var isPointer, isInterface, isArray, isSlice, isMap, isStruct, isError bool
	var fieldPkgPath, fieldPkg, fieldType string
	var mapKey, mapValue *Type
	var structFields map[string]StructField
	var interfaceMethods map[string]Func

	switch t := typ.(type) {
	case *types.Interface:
		isInterface = true
		interfaceMethods = ExtractInterfaceMethods(t)
	case *types.Pointer:
		isPointer = true
		typ = t.Elem()
	}

	switch t := typ.(type) {
	case *types.Slice:
		isSlice = true
		typ = t.Elem()
	case *types.Array:
		isArray = true
		typ = t.Elem()
	case *types.Map:
		isMap = true
		mapKey = NewType(t.Key())
		mapValue = NewType(t.Elem())
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
		if pkg := obj.Pkg(); pkg != nil {
			fieldPkg = pkg.Name()
			fieldPkgPath = pkg.Path()
		}
		fieldType = obj.Name()

		// The underlying type could be a struct.
		if structType, ok := obj.Type().Underlying().(*types.Struct); ok {
			isStruct = true
			structFields = ExtractStructFields(structType)
		}

		// The underlying type could be a interface.
		if types.IsInterface(obj.Type().Underlying()) {
			isInterface = true
			interfaceMethods = ExtractInterfaceMethods(obj.Type().Underlying().(*types.Interface))
		}
	case *types.Struct:
		isStruct = true
		structFields = ExtractStructFields(t)
	default:
		fieldType = t.String()
	}

	isError = fieldType == "error"
	return &Type{
		Type:             fieldType,
		Pkg:              fieldPkg,
		PkgPath:          fieldPkgPath,
		IsStruct:         isStruct,
		IsSlice:          isSlice,
		IsArray:          isArray,
		IsPointer:        isPointer,
		IsMap:            isMap,
		IsInterface:      isInterface,
		IsError:          isError,
		MapKey:           mapKey,
		MapValue:         mapValue,
		StructFields:     structFields,
		InterfaceMethods: interfaceMethods,
		T:                typ,
	}
}

// StructField for the example below.
//type Foo struct {
//  Name sql.NullString `json:"name"
//}
type StructField struct {
	Name string `example:"Name"`
	// Useful when the output directory doesn't match the existing ones.
	PkgPath  string // e.g. github.com/your-org/yourpkg
	PkgName  string // e.g. yourpkg
	Exported bool   // e.g. true
	Tag      *Tag   // e.g. `map:"RenameField,CustomFunction"`
	*Type
	*types.Var
}
