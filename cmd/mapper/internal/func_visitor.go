package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type FuncVisitor struct {
	Param  *FuncParamVisitor
	Result *FuncResultVisitor
}

func (f *FuncVisitor) Visit(fn *types.Func) {
	/*
		func (m Mapper) mapFooToBar(f0 Foo) Bar {
			return Bar{
				Field0: f0.Field0,
				Field1: f0.Field1(), 											// Method call.
				Field2: customFunc(f0.Field2()), 					// Custom func.
				Field3: m.struct.method(f0.Field3()), 		// Custom struct method.
				Field4: m.interface.method(f0.Field4()), 	// Custom interface method.
			}
		}

		func (m Mapper) mapFooToBarWithError(f0 Foo) (Bar, error) {
			// Each of them may have error handling too.
			f0Field1, err := f0.Field1()
			if err != nil {
				return Bar{}, err
			}

			f0Field2, err := customFunc(f0.Field2())
			if err != nil {
				return Bar{}, err
			}

			f0Field3, err := m.struct.method(f0.Field3())
			if err != nil {
				return Bar{}, err
			}

			f0Field4, err := m.interface.method(f0.Field4())
			if err != nil {
				return Bar{}, err
			}

			return Bar{
				Field0: f0.Field0,
				Field1: f0Field1,  // Method call.
				Field2: f0Field2,  // Custom func.
				Field3: f0Field3,  // Custom struct method.
				Field4: f0Field4,  // Custom interface method.
			}, nil
		}

		// Each method also supports slice to slice conversion.
	*/
	sig := fn.Type().(*types.Signature)
	// Parse args
	// Parse result

	// Check if A -> B can be mapped
	// - must be struct/slice
	// Load all struct field tags
	// struct field return methods must match the rhs field return type.
	// struct field input must match all the lhs input

	// checkFuncHasOneParam
	npar := sig.Params().Len()
	if npar != 1 {
		panic("invalid param count")
	}

	// checkFuncHasOneResult
	nres := sig.Results().Len()
	if nres < 1 || nres > 2 {
		panic("invalid result count")
	}

	var hasError bool
	if nres > 1 {
		if !mapper.IsUnderlyingError(sig.Results().At(1).Type()) {
			panic("tuple return must be error")
		}
		hasError = true
	}

	param := sig.Params().At(0).Type()
	paramVisitor := NewFuncParamVisitor()
	_ = mapper.Walk(paramVisitor, param)

	result := sig.Results().At(0).Type()
	resultVisitor := NewFuncResultVisitor()
	_ = mapper.Walk(resultVisitor, result)

	/*
		The custom function loaded has error, but parent does not have.

		func (m *Mapper) MapFooToBar(f0 Foo) Bar {
			b, err := customFunc(f0)
			if err != nil {
				// There are no error return ...
			}
			return b
		}

	*/
	// checkFuncMissingError
	if resultVisitor.HasError() && !hasError {
		// Invalid error signature
		panic(PrettyError(`
			function %q is missing error return
			hint: add error return
		`, PrettyFuncSignature(fn)))
	}

	// checkTypesMatchs
	if paramVisitor.isCollection != resultVisitor.isCollection {
		if paramVisitor.isCollection {
			panic(PrettyError(`
				invalid function %q
				hint: cannot map slice to non-slice
			`, PrettyFuncSignature(fn)))
		} else {
			panic(PrettyError(`
				invalid function %q
				hint: cannot map non-slice to slice
			`, PrettyFuncSignature(fn)))
		}
	}

	f.Result = resultVisitor
	f.Param = paramVisitor
}

func (f FuncVisitor) HasError() bool {
	return f.Result.HasError() || f.Param.HasError()
}
