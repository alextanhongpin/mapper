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
		v.methods = mapper.NewInterfaceMethods(u)
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

			var lhsType types.Type
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
				paramType := mapperFn.From.Type
				resultType := mapperFn.To.Type

				if !mapper.IsUnderlyingIdentical(lhsType, paramType) {
					panic("input type does not match func arg")
				}

				if !mapper.IsUnderlyingIdentical(rhsType, resultType) {
					panic("output type does not match func result")
				}

				// Type already matches, continue.
				continue
			}

			if !mapper.IsUnderlyingIdentical(lhsType, rhsType) {
				innerSignature := mapper.NewFunc(mapper.NormFuncFromTypes("", lhsType, rhsType)).Signature()
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

func (v *InterfaceVisitor) MethodInfo(name string) (*FuncVisitor, bool) {
	info, ok := v.methodInfo[name]
	return info, ok
}
