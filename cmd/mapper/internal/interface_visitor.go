package internal

import (
	"fmt"
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type InterfaceVisitor struct {
	methods                           map[string]*mapper.Func
	methodInfo                        map[string]*FuncVisitor
	generatedByPrivateMethodSignature map[string]bool
	hasErrorByPrivateMethod           map[string]bool
}

func (v *InterfaceVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Interface:
		v.methods = mapper.ExtractInterfaceMethods(u)
		v.parseMethods()
	}
	return true
}

func NewInterfaceVisitor(T types.Type) *InterfaceVisitor {
	v := &InterfaceVisitor{
		generatedByPrivateMethodSignature: make(map[string]bool),
		hasErrorByPrivateMethod:           make(map[string]bool),
		methodInfo:                        make(map[string]*FuncVisitor),
	}
	_ = mapper.Walk(v, T.Underlying())
	return v
}

func GenerateInterfaceMethods(T types.Type) map[string]*mapper.Func {
	v := &InterfaceVisitor{}
	_ = mapper.Walk(v, T.Underlying())
	return v.methods
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
		for name, rhs := range result.Fields() {
			_, hasField := param.Fields()[name]
			_, hasMethod := param.Methods()[name]

			if !(hasField || hasMethod) {
				panic(fmt.Errorf("no mapping found for %q", name))
			}

			// There's a custom mapper.
			if rhs.Tag != nil && rhs.Tag.HasFunc() {
				//mapperFn(lhs) rhs
				mapperFn := result.MappersByTag()[rhs.Tag.Name]
				if mapperFn.Error {
					v.hasErrorByPrivateMethod[signature] = true
				}
			}
		}
		v.generatedByPrivateMethodSignature[signature] = false
	}

	for name := range v.methods {
		res := v.methodInfo[name]

		result, param := res.Result, res.Param

		for name, rhs := range result.Fields() {
			field, isField := param.Fields()[name]
			method := param.Methods()[name]

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
				mapperFn := result.MappersByTag()[rhs.Tag.Name]
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
				if !v.generatedByPrivateMethodSignature[innerSignature] {
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

func (v *InterfaceVisitor) GeneratedByPrivateMethodSignature() map[string]bool {
	return v.generatedByPrivateMethodSignature
}

func (v *InterfaceVisitor) HasErrorByPrivateMethod() map[string]bool {
	return v.hasErrorByPrivateMethod
}
