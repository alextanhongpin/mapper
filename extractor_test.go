package mapper_test

import (
	"testing"

	"github.com/alextanhongpin/mapper"
	"github.com/google/go-cmp/cmp"
)

func TestSignature(t *testing.T) {
	m := mapper.Func{
		Name: "Convert",
		From: &mapper.FuncArg{
			Name: "f",
			Type: &mapper.Type{
				Type:     "Foo",
				IsStruct: true,
			},
		},
		To: &mapper.FuncArg{
			Name: "b",
			Type: &mapper.Type{
				Type:     "Bar",
				IsStruct: true,
			},
		},
		Error: &mapper.Type{
			Type:    "error",
			IsError: true,
		},
	}

	expected := "func mapFooToBar(Foo) (Bar, error)"
	got := m.NormalizedSignature()
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Fatalf("(want +, got -): \n%s", diff)
	}
}
