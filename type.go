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

func (t Type) Normalize() *Type {
	return &Type{
		Type:         t.Type,
		Pkg:          t.Pkg,
		PkgPath:      t.PkgPath,
		StructFields: t.StructFields,
		T:            t.T,
	}
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

// EqualElem checks if the type a.A, regardless of whether
// it is pointer, slice etc, matches type b.B.
// The elem is only considered the same if they reside in
// the same pkg.
// So a.A is not the same as b.A even if both A has same types.
func (t Type) EqualElem(other *Type) bool {
	return t.Type == other.Type && t.PkgPath == other.PkgPath
}
