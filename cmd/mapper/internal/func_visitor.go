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
	// checkFuncMissingError
	if resultVisitor.hasError && !hasError {
		// Invalid error signature
		panic("error not implemented")
	}

	// checkFieldsHasMappings
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
			// NOTE: Load all function before doing this checking.
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
	// Deferred
	// checkTypesMatchs
	if paramVisitor.isCollection != resultVisitor.isCollection {}
}
type Transform struct {
	Left, Right interface{}
}

func VisitFuncs(fns map[string]*mapper.Func) {
	// privateMethods := make(map
	// infos ...

	// for each fn
	//    info := VisitFunc(fn)
	//    store the info for later used
	//    privateMethods[info.normalizedSignature] = false
	
	// full validation can now be done
	// for each info
	//     validate each field
	//         rhs field must match lgs field or method
	//         if got tag, output of fn must match rhs field, input of fn must match lhs field or method
	//         if no tag, it must match one of the privatemethods, or type must be equal, except pointer
	
	// Time to build private methods
	// is method? expand the method first
	// has tag? apply transformation
	// has private method? apply it
	// pointer? apply it
}

func IsUnderlyingIdentical() bool {
    return false
}
