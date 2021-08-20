package mapper

import (
	"go/types"
)

type Type struct {
	Type             string `example:"NullString"`
	PkgPath          string `example:"database/sql"`
	IsStruct         bool
	IsPointer        bool
	IsArray          bool // Whether it's an array or slice.
	IsInterface      bool
	IsSlice          bool
	IsError          bool // NOT IMPLEMENTED
	IsMap            bool
	MapKey           *Type
	MapValue         *Type
	StructFields     map[string]StructField
	StructMethods    []Func
	InterfaceMethods []Func
	T                types.Type
}

// NewType recursively checks for the field type.
func NewType(typ types.Type) *Type {
	var isPointer, isInterface, isArray, isSlice, isMap, isStruct, isError bool
	var fieldPkgPath, fieldType string
	var mapKey, mapValue *Type
	var structFields map[string]StructField
	var structMethods, interfaceMethods []Func

	switch t := typ.(type) {
	case *types.Interface:
		isInterface = true
		interfaceMethods = extractInterfaceMethods(t)
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
			fieldPkgPath = pkg.Path()
		}
		fieldType = obj.Name()
		structMethods = extractNamedMethods(t)

		// The underlying type could be a struct.
		if structType, isStruct := t.Underlying().(*types.Struct); isStruct {
			isStruct = true
			structFields = extractStructFields(structType)
		}
	case *types.Struct:
		isStruct = true
		structFields = extractStructFields(t)
	default:
		fieldType = t.String()
	}

	isError = fieldType == "error"
	return &Type{
		Type:             fieldType,
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
		StructMethods:    structMethods,
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
	PkgPath  string `example:"github.com/alextanhongpin/go-codegen/test"`
	PkgName  string `example:"test"`
	Exported bool   `example:"true"`
	Tag      string `example:"build:'-'"` // To ignore builder.
	*Type
}
