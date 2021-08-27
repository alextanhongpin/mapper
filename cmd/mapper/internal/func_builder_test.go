package internal_test

import (
	"strings"
	"testing"

	"github.com/alextanhongpin/mapper"
	"github.com/alextanhongpin/mapper/cmd/mapper/internal"
	"github.com/google/go-cmp/cmp"
)

func newStructType(structName string, isStructFieldPointer bool) *mapper.Type {
	// Output:
	// type StructName struct {
	//   Name string
	//   Age int64
	// }
	structType := &mapper.Type{
		Type:     structName,
		IsStruct: true,
		StructFields: map[string]mapper.StructField{
			"Name": {
				Name: "Name",
				Type: &mapper.Type{
					Type:      "string",
					IsPointer: isStructFieldPointer,
				},
			},
			"Age": {
				Name: "Age",
				Type: &mapper.Type{
					Type:      "int",
					IsPointer: isStructFieldPointer,
				},
			},
		},
	}
	return structType
}

func block(code string) string {
	return strings.TrimSpace(code)
}

type generateArgs struct {
	hasError                   bool
	isInputPointer             bool
	isOutputPointer            bool
	isInputStructFieldPointer  bool
	isOutputStructFieldPointer bool
}

func generate(args generateArgs) string {
	hasError := args.hasError
	isInputPointer := args.isInputPointer
	isOutputPointer := args.isOutputPointer
	isInputStructFieldPointer := args.isInputStructFieldPointer
	isOutputStructFieldPointer := args.isOutputStructFieldPointer

	field := "Name"
	structA := newStructType("A", isInputStructFieldPointer)
	structB := newStructType("B", isOutputStructFieldPointer)

	var customFuncErr *mapper.Type
	if hasError {
		customFuncErr = &mapper.Type{Type: "error"}
	}

	fb := internal.NewFuncBuilder(
		internal.NewFieldResolver("a", structA.StructFields[field], structB.StructFields[field]),
		// The parent func.
		// func (m *Mapper) mapAtoB(a A) (b B, err error) {
		//   // Do sth
		// }
		&mapper.Func{
			Name:  "mapAtoB",
			From:  mapper.NewFuncArg("a", structA, false),
			To:    mapper.NewFuncArg("b", structB, false),
			Error: customFuncErr, // Parent must have error too.
		},
	)

	// The local function.
	customFunc := &mapper.Func{
		Name:  "stringToString",
		From:  mapper.NewFuncArg("a", &mapper.Type{Type: "string", IsPointer: isInputPointer}, false),
		To:    mapper.NewFuncArg("b", &mapper.Type{Type: "string", IsPointer: isOutputPointer}, false),
		Error: customFuncErr,
	}

	var c internal.C
	fb.BuildFuncCall(&c, customFunc, structA.StructFields[field].Type, structB.StructFields[field].Type)
	return c.String()
}

func TestGenBuilder(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		code := generate(generateArgs{})
		if diff := cmp.Diff("a0Name := stringToString(a0.Name)", code); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("basic input pointer", func(t *testing.T) {
		t.Run("ANamePtr=false, BNamePtr=false", func(t *testing.T) {
			code := generate(generateArgs{isInputPointer: true})
			if diff := cmp.Diff(block(`
a0Name := stringToString(&a0.Name)
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=true, BNamePtr=false", func(t *testing.T) {
			code := generate(generateArgs{isInputPointer: true, isInputStructFieldPointer: true})
			if diff := cmp.Diff(block(`
var a0Name string
if a0.Name != nil {
	a0Name = stringToString(a0.Name)
}
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=false, BNamePtr=true", func(t *testing.T) {
			code := generate(generateArgs{isInputPointer: true, isOutputStructFieldPointer: true})
			if diff := cmp.Diff(block(`
a0Name := stringToString(&a0.Name)
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=true, BNamePtr=true", func(t *testing.T) {
			code := generate(generateArgs{
				isInputPointer:             true,
				isInputStructFieldPointer:  true,
				isOutputStructFieldPointer: true,
			})
			if diff := cmp.Diff(block(`
var a0Name *string
if a0.Name != nil {
	tmp := stringToString(a0.Name)
	a0Name = &tmp
}
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
	})

	t.Run("basic output pointer", func(t *testing.T) {
		t.Run("ANamePtr=false, BNamePtr=false", func(t *testing.T) {
			code := generate(generateArgs{isOutputPointer: true})
			if diff := cmp.Diff(block(`
a0Name := stringToString(a0.Name)
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=true, BNamePtr=false", func(t *testing.T) {
			code := generate(generateArgs{isOutputPointer: true, isInputStructFieldPointer: true})
			if diff := cmp.Diff(block(`
var a0Name string
if a0.Name != nil {
	tmp := stringToString(*a0.Name)
	if tmp != nil {
		a0Name = *tmp
	}
}
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=false, BNamePtr=true", func(t *testing.T) {
			code := generate(generateArgs{isOutputPointer: true, isOutputStructFieldPointer: true})
			if diff := cmp.Diff(block(`
a0Name := stringToString(a0.Name)
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=true, BNamePtr=true", func(t *testing.T) {
			code := generate(generateArgs{
				isOutputPointer:            true,
				isInputStructFieldPointer:  true,
				isOutputStructFieldPointer: true,
			})
			if diff := cmp.Diff(block(`
var a0Name *string
if a0.Name != nil {
	a0Name = stringToString(*a0.Name)
}
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
	})

	t.Run("basic input and output pointer", func(t *testing.T) {
		t.Run("ANamePtr=false, BNamePtr=false", func(t *testing.T) {
			code := generate(generateArgs{
				isInputPointer:             true,
				isOutputPointer:            true,
				isInputStructFieldPointer:  false,
				isOutputStructFieldPointer: false,
			})
			if diff := cmp.Diff(block(`
a0Name := stringToString(&a0.Name)
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=true, BNamePtr=false", func(t *testing.T) {
			code := generate(generateArgs{
				isInputPointer:             true,
				isOutputPointer:            true,
				isInputStructFieldPointer:  true,
				isOutputStructFieldPointer: false,
			})
			if diff := cmp.Diff(block(`
var a0Name string
if a0.Name != nil {
	tmp := stringToString(a0.Name)
	if tmp != nil {
		a0Name = *tmp
	}
}
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=false, BNamePtr=true", func(t *testing.T) {
			code := generate(generateArgs{
				isInputPointer:             true,
				isOutputPointer:            true,
				isInputStructFieldPointer:  false,
				isOutputStructFieldPointer: true,
			})
			if diff := cmp.Diff(block(`
a0Name := stringToString(&a0.Name)
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
		t.Run("ANamePtr=true, BNamePtr=true", func(t *testing.T) {
			code := generate(generateArgs{
				isInputPointer:             true,
				isOutputPointer:            true,
				isInputStructFieldPointer:  true,
				isOutputStructFieldPointer: true,
			})
			if diff := cmp.Diff(block(`
var a0Name *string
if a0.Name != nil {
	a0Name = stringToString(a0.Name)
}
`), code); diff != "" {
				t.Fatal(diff)
			}
		})
	})

	t.Run("with error", func(t *testing.T) {
		code := generate(generateArgs{hasError: true})
		if diff := cmp.Diff(block(`
a0Name, err := stringToString(a0.Name)
if err != nil {
	return B{}, err
}
`), code); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("with error and input pointer", func(t *testing.T) {
		code := generate(generateArgs{hasError: true, isInputPointer: true})
		if diff := cmp.Diff(block(`
a0Name, err := stringToString(&a0.Name)
if err != nil {
	return B{}, err
}
`), code); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("with error and output pointer", func(t *testing.T) {
		code := generate(generateArgs{hasError: true, isOutputPointer: true})
		if diff := cmp.Diff(block(`
a0Name, err := stringToString(a0.Name)
if err != nil {
	return B{}, err
}
`), code); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("with error and input and output pointer", func(t *testing.T) {
		code := generate(generateArgs{hasError: true, isInputPointer: true, isOutputPointer: true})
		if diff := cmp.Diff(block(`
a0Name, err := stringToString(&a0.Name)
if err != nil {
	return B{}, err
}
`), code); diff != "" {
			t.Fatal(diff)
		}
	})
}
