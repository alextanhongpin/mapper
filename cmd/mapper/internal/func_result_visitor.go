package internal

import (
	"fmt"
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type FuncResultVisitor struct {
	fields       mapper.StructFields
	mappersByTag map[string]*mapper.Func
	isCollection bool
	isPointer    bool
	obj          *types.TypeName
}

func NewFuncResultVisitor() *FuncResultVisitor {
	return &FuncResultVisitor{
		mappersByTag: make(map[string]*mapper.Func),
	}
}

func (v *FuncResultVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Pointer:
		v.isPointer = true
	case *types.Slice, *types.Array:
		v.isCollection = true
	case *types.Named:
		v.obj = u.Obj()
		return true
	case *types.Struct:
		v.fields = mapper.ExtractStructFields(u)
		for _, field := range v.fields.WithTags() {
			tag := field.Tag
			if tag == nil {
				continue
			}
			if !tag.HasFunc() {
				continue
			}
			var m *mapper.Func
			if tag.IsFunc() {
				fn := loadFunc(field)
				v.mappersByTag[tag.Name] = fn
				// Avoid overwriting if the struct loads multiple
				// functions and some does not have errors.
				m = fn

			}
			if tag.IsMethod() {
				met := loadMethod(field)
				v.mappersByTag[tag.Name] = met
				m = met
			}

			/*
				Return underlying type should match.
				If the field type is []A, *A or just A, then the function/method should
				also return the equivalent type A.

				type Foo struct {
					// AddSalutation accepts string, so it will be mapped over the names.
					names []string `map:",AddSalutation"`

					// CheckAge returns an error as the second return value.
					age int64 `map:",CheckAge"`
				}

				func Rename(name string) string {
					return "Mr/Ms " + name
				}

				func CheckAge(age int64) (int64, error) {
					if age < 0 || age > 150 {
						return 0, errors.New("invalid age")
					}
					return age, nil
				}
			*/
			if !types.IdenticalIgnoreTags(
				NewUnderlyingType(m.To.Type.E),
				NewUnderlyingType(field.E),
			) {
				panic("not equal type")
			}
		}
	default:
		panic("not implemented")
	}
	return false
}

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

	fnType, ok := obj.(*types.Func)
	if !ok {
		panic(fmt.Sprintf("mapper: %q is not a func", tag.Func))
	}

	return mapper.NewFunc(fnType)
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

	if _, ok := obj.Type().(*types.Named); !ok {
		panic(fmt.Errorf("tag %q is invalid\ndetail: %q is not a struct or interface", tag.Tag, tag.TypeName))
	}

	T := obj.Type().Underlying()
	if types.IsInterface(T) {
		interfaceMethods := GenerateInterfaceMethods(T)
		return interfaceMethods[tag.Func]
	}

	structMethods := mapper.ExtractNamedMethods(T)
	return structMethods[tag.Func]
}

func (v *FuncResultVisitor) HasError() bool {
	for _, fn := range v.mappersByTag {
		if fn.Error {
			return true
		}
	}
	return false
}

func (v FuncResultVisitor) Fields() mapper.StructFields {
	return v.fields
}

func (v FuncResultVisitor) MappersByTag() map[string]*mapper.Func {
	return v.mappersByTag
}

func (v FuncResultVisitor) IsCollection() bool {
	return v.isCollection
}

func (v FuncResultVisitor) IsPointer() bool {
	return v.isPointer
}
