package internal_test

import (
	"fmt"
	"testing"

	"github.com/alextanhongpin/mapper"
	"github.com/alextanhongpin/mapper/cmd/mapper/internal"
	"github.com/google/go-cmp/cmp"
)

func TestGenBuilder(t *testing.T) {

	stringType := &mapper.Type{
		Type: "string",
	}

	newStructType := func(structName string) *mapper.Type {
		structType := &mapper.Type{
			Type:     structName,
			IsStruct: true,
			StructFields: map[string]mapper.StructField{
				"a": {
					Name: "A",
					Type: stringType,
				},
			},
		}
		return structType
	}
	structA := newStructType("A")
	structB := newStructType("B")

	fb := internal.NewFuncBuilder(
		internal.NewFieldResolver("a", mapper.StructField{
			Name: "A",
			Type: stringType,
		}, mapper.StructField{
			Name: "A",
			Type: stringType,
		}),
		// The parent func.
		&mapper.Func{
			Name: "MapAtoB",
			From: mapper.NewFuncArg("a", structA, false),
			To:   mapper.NewFuncArg("b", structB, false),
			//Error: &mapper.Type{},
		},
	)

	stringToString := &mapper.Func{
		Name: "stringToString",
		From: mapper.NewFuncArg("a", structA, false),
		To:   mapper.NewFuncArg("b", structB, false),
		//Error: &mapper.Type{},
	}

	var c internal.C
	fb.BuildFuncCall(&c, stringToString, stringType, stringType)
	fb.BuildFuncCall(&c, stringToString, stringType, stringType)

	if diff := cmp.Diff("a0A := stringToString(a0.A)", fmt.Sprintf("%#v", c[0])); diff != "" {
		t.Fatal(diff)
	}
	if diff := cmp.Diff("a1A := stringToString(a0A)", fmt.Sprintf("%#v", c[1])); diff != "" {
		t.Fatal(diff)
	}
}
