package internal

import (
	"fmt"
	"go/types"

	"github.com/alextanhongpin/mapper"
)

func loadFunc(field mapper.StructField) *mapper.Func {
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
		panic("func not found")
	}

	T, ok := obj.(*types.Func)
	if !ok {
		panic(fmt.Sprintf("mapper: %q is not a func", tag.Func))
	}

	return mapper.NewFunc(T)
}

func loadMethod(field mapper.StructField) *mapper.Func {
	tag := field.Tag
	fieldPkgPath := field.PkgPath
	if tag.IsImported() {
		fieldPkgPath = tag.PkgPath
	}

	// Load the interface/struct.
	pkg := mapper.LoadPackage(fieldPkgPath)
	obj := mapper.LookupType(pkg, tag.TypeName)
	if obj == nil {
		panic(fmt.Errorf("tag %q is invalid\ndetail: %q not found\nhelp: check if the type %q exists", tag.Tag, tag.TypeName, tag.TypeName))
	}

	named, ok := obj.Type().(*types.Named)
	if !ok {
		panic(fmt.Errorf("tag %q is invalid\ndetail: %q is not a struct or interface", tag.Tag, tag.TypeName))
	}

	T := obj.Type()
	if types.IsInterface(T) {
		interfaceMethods := GenerateInterfaceMethods(T)
		fn := interfaceMethods[tag.Func]
		fn.Obj = named.Obj()
		return fn
	}

	structMethods := mapper.ExtractNamedMethods(T)
	method := structMethods[tag.Func]
	method.Obj = named.Obj()
	return method
}
