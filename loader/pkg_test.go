package loader_test

import (
	"go/types"
	"testing"

	"github.com/alextanhongpin/mapper/loader"
)

var program = `
package main


type User struct {
	Name string
	Age int64
}

func main() {}
`

func TestLookupStruct(t *testing.T) {
	pkg := loader.LoadPackageString(program)

	obj := pkg.Scope().Lookup("User")

	T := obj.Type()
	_, ok := T.(*types.Named)
	if !ok {
		t.Fatalf("expected type to be named, got %v", T)
	}
	U := T.Underlying()
	_, ok = U.(*types.Struct)
	if !ok {
		t.Fatalf("expected type to be struct, got %v", U)
	}
}
