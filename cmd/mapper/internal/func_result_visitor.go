package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type FuncResultVisitor struct {
	fields       mapper.StructFields
	mappersByTag map[string]*mapper.Func
	isCollection bool
}

func NewFuncResultVisitor() *FuncResultVisitor {
	return &FuncResultVisitor{
		mappersByTag: make(map[string]*mapper.Func),
	}
}

func (v *FuncResultVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Array, *types.Slice:
		v.isCollection = true
	case *types.Named:
	case *types.Struct:
		v.fields = mapper.NewStructFields(u).WithTags()
		for _, field := range v.fields {
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
				v.mappersByTag[tag.Tag] = fn
				// Avoid overwriting if the struct loads multiple
				// functions and some does not have errors.
				m = fn
			}

			if tag.IsMethod() {
				met := loadMethod(field)
				v.mappersByTag[tag.Tag] = met
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
			if !mapper.IsUnderlyingIdentical(m.To.Type, field.Type) {
				panic("not equal type")
			}
		}
	}
	return true
}

func (v *FuncResultVisitor) HasError() bool {
	for _, fn := range v.mappersByTag {
		if fn.Error {
			return true
		}
	}
	return false
}

func (v FuncResultVisitor) Fields() []string {
	var result []string
	for name := range v.fields {
		result = append(result, name)
	}
	return result
}

func (v FuncResultVisitor) FieldByName(name string) (mapper.StructField, bool) {
	field, ok := v.fields[name]
	return field, ok
}

func (v FuncResultVisitor) MapperByTag(tag string) (*mapper.Func, bool) {
	mapper, ok := v.mappersByTag[tag]
	return mapper, ok
}
