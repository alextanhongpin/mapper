package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

func VisitFunc(fn *types.Func) {
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
	npar := sig.Params().Len()
	if npar != 1 {
		panic("invalid param count")
	}

	nres := sig.Results().Len()
	if nres < 1 || nres > 2 {
		panic("invalid result count")
	}

	var hasError bool
	if nres > 1 {
		if !IsUnderlyingError(sig.Results().At(1).Type()) {
			panic("tuple return must be error")
		}
		hasError = true
	}

	paramVisitor := NewFuncParamVisitor()
	_ = mapper.Walk(paramVisitor, sig.Params().At(0).Type())

	resultVisitor := NewFuncResultVisitor()
	_ = mapper.Walk(resultVisitor, sig.Results().At(0).Type())

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
	if resultVisitor.hasError && !hasError {
		// Invalid error signature
		panic("error not implemented")
	}

	for name, rhs := range resultVisitor.fields {
		// It's a field mapping.
		if lhs, ok := paramVisitor.fields[name]; ok {
			if tagFn, ok := resultVisitor.mappersByTag[rhs.Tag.Name]; ok {
				// There tag fn input does not match.
				if !tagFn.From.Type.EqualElem(lhs.Type) {
					panic("invalid input type")
				}
				// The tag function input matches.
				continue
			}
			// TODO: Check private mapper method.
			if !lhs.Type.EqualElem(rhs.Type) {
				panic("cannot map field")
			}
			continue
		}

		if lhs, ok := paramVisitor.methods[name]; ok {
			if !lhs.To.Type.EqualElem(rhs.Type) {
				panic("cannot map field")
			}
			continue
		}
		panic("not matching field found")
	}
}
