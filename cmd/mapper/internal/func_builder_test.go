package internal_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	_ "embed"

	"github.com/alextanhongpin/mapper"
	"github.com/alextanhongpin/mapper/cmd/mapper/internal"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

var (
	//go:embed testdata/1_argptr.yaml
	argsPointerData string

	//go:embed testdata/2_argptr_resptr.yaml
	argsAndResultPointerData string

	//go:embed testdata/3_has_error.yaml
	hasErrorData string

	//go:embed testdata/4_has_error_argptr.yaml
	hasErrorArgsPointerData string

	//go:embed testdata/5_has_error_resptr.yaml
	hasErrorResultPointerData string

	//go:embed testdata/6_has_error_argptr_resptr.yaml
	hasErrorArgsAndResultPointerData string
)

type testdata struct {
	Args     generateArgs `yaml:"args"`
	Expected string       `yaml:"expected"`
}

type generateArgs struct {
	HasError        bool `yaml:"error"`
	IsInputPointer  bool `yaml:"inPtr"`
	IsOutputPointer bool `yaml:"outPtr"`
	IsArgPointer    bool `yaml:"argPtr"`
	IsResultPointer bool `yaml:"resPtr"`
	Skip            bool `yaml:"skip"`
}

func (a generateArgs) Scenario(pos int) string {
	return fmt.Sprintf("test %d.error:%t,argptr:%t,resptr:%t,inptr:%t,outptr:%t",
		pos,
		a.HasError,
		a.IsArgPointer,
		a.IsResultPointer,
		a.IsInputPointer,
		a.IsOutputPointer,
	)
}

func TestGenBuilder(t *testing.T) {
	testGroup(t, "arg pointer", argsPointerData)
	testGroup(t, "arg and res pointer", argsAndResultPointerData)
	testGroup(t, "has error", hasErrorData)
	testGroup(t, "has error arg pointer", hasErrorArgsPointerData)
	testGroup(t, "has error res pointer", hasErrorResultPointerData)
	testGroup(t, "has error arg and res pointer", hasErrorArgsAndResultPointerData)
}

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

func generate(args generateArgs) string {
	hasError := args.HasError
	isArgPointer := args.IsArgPointer
	isResultPointer := args.IsResultPointer
	isInputPointer := args.IsInputPointer
	isOutputPointer := args.IsOutputPointer

	field := "Name"
	structA := newStructType("A", isInputPointer)
	structB := newStructType("B", isOutputPointer)

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
		From:  mapper.NewFuncArg("a", &mapper.Type{Type: "string", IsPointer: isArgPointer}, false),
		To:    mapper.NewFuncArg("b", &mapper.Type{Type: "string", IsPointer: isResultPointer}, false),
		Error: customFuncErr,
	}

	var c internal.C
	fb.BuildFuncCall(&c, customFunc, structA.StructFields[field].Type, structB.StructFields[field].Type)
	return c.String()
}

func block(code string) string {
	return spaceToTab(strings.TrimSpace(code))
}

// yaml does not support tabs,
func spaceToTab(input string) string {
	return strings.ReplaceAll(input, "  ", "\t")
}

func testGroup(t *testing.T, name, raw string) {
	t.Run(name, func(t *testing.T) {

		parts := strings.Split(raw, "---")
		for i, part := range parts {
			var data testdata
			err := yaml.Unmarshal([]byte(part), &data)
			if err != nil {
				log.Fatalf("cannot unmarshal data: %v", err)
			}

			t.Run(data.Args.Scenario(i+1), func(t *testing.T) {
				if data.Args.Skip {
					t.Skip()
				}

				code := generate(data.Args)
				if diff := cmp.Diff(block(data.Expected), code); diff != "" {
					t.Fatal(diff)
				}
			})
		}
	})
}
