package internal

import (
	"fmt"
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type InterfaceVisitor struct {
	methods          map[string]*mapper.Func
	methodInfo       map[string]*FuncVisitor
	mappers          map[string]bool
	hasErrorByMapper map[string]bool
}

func (v *InterfaceVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Interface:
		v.methods = mapper.ExtractInterfaceMethods(u)
		v.parseMethods()
		return false
	}
	return true
}

func NewInterfaceVisitor(T types.Type) *InterfaceVisitor {
	v := &InterfaceVisitor{
		mappers:          make(map[string]bool),
		hasErrorByMapper: make(map[string]bool),
		methodInfo:       make(map[string]*FuncVisitor),
	}
	_ = mapper.Walk(v, T.Underlying())
	return v
}

func GenerateInterfaceMethods(T types.Type) map[string]*mapper.Func {
	u := T.Underlying().(*types.Interface)
	return mapper.ExtractInterfaceMethods(u)
}

func (v *InterfaceVisitor) parseMethods() {
	for name, fn := range v.methods {
		iv := &FuncVisitor{}
		iv.Visit(fn.Fn)
		// Store the func info.
		v.methodInfo[name] = iv
		signature := fn.Normalize().Signature()

		result, param := iv.Result, iv.Param

		// checkFieldsHasMappings
		for _, name := range result.Fields() {
			rhs, _ := result.FieldByName(name)
			_, hasField := param.FieldByName(name)
			_, hasMethod := param.MethodByName(name)

			if !(hasField || hasMethod) {
				panic(fmt.Errorf("no mapping found for %q", name))
			}

			// There's a custom mapper.
			if rhs.Tag != nil && rhs.Tag.HasFunc() {
				//mapperFn(lhs) rhs
				mapperFn, _ := result.MapperByTag(rhs.Tag.Tag)
				if mapperFn.Error {
					v.hasErrorByMapper[signature] = true
				}
			}
		}
		v.mappers[signature] = true
	}

	for name := range v.methods {
		res := v.methodInfo[name]

		result, param := res.Result, res.Param

		for _, name := range result.Fields() {
			rhs, _ := result.FieldByName(name)
			field, isField := param.FieldByName(name)
			method, _ := param.MethodByName(name)

			var lhsType *mapper.Type
			if isField {
				lhsType = field.Type
			} else {
				lhsType = method.To.Type
			}
			rhsType := rhs.Type

			// There's a custom mapper.
			if rhs.Tag != nil && rhs.Tag.HasFunc() {
				/*
					func CustomFunc(param Param) (Result) {
					}

					type LHS struct {
						param Param
					}

					type RHS struct {
						result Result
					}

					CustomFunc(LHS.param) == RHS.result

				*/
				mapperFn, _ := result.MapperByTag(rhs.Tag.Tag)
				paramType := mapperFn.From.Type.T
				resultType := mapperFn.To.Type.T

				if !IsUnderlyingIdentical(lhsType.T, paramType) {
					panic("input type does not match func arg")
				}

				if !IsUnderlyingIdentical(rhsType.T, resultType) {
					panic("output type does not match func result")
				}

				// Type already matches, continue.
				continue
			}

			if !IsUnderlyingIdentical(lhsType.T, rhsType.T) {
				innerSignature := mapper.NewFunc(mapper.NormFuncFromTypes(lhsType, rhsType)).Signature()
				if !v.mappers[innerSignature] {
					panic("no conversion found for field")
				}
			}
		}
	}
}

func (v *InterfaceVisitor) Methods() map[string]*mapper.Func {
	return v.methods
}

func (v *InterfaceVisitor) MethodInfo() map[string]*FuncVisitor {
	return v.methodInfo
}

func (v *InterfaceVisitor) MappersByName() map[string]bool {
	return v.mappers
}
