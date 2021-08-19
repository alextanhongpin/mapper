package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/alextanhongpin/mapper"
	"github.com/alextanhongpin/mapper/examples/bar"
	"github.com/alextanhongpin/mapper/examples/foo"
	"github.com/dave/jennifer/jen"
	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
)

const Generator = "mapper"

type Task struct {
	Name string
}

type Foo struct {
	Age  int
	name string
	Task Task
	ID   uuid.UUID
	//Remarks string // Fail with extra fields now.
	// CustomName `mapper:"YourName"`
	// CustomMapper `mapper:"github.com/yourorganization/yourpackage/struct.Method"`
	// CustomInterface `mapper:"github.com/yourorganization/yourpackage/interface.Method"`
	// CustomFunction`mapper:"github.com/yourorganization/yourpackage.funcName"`
}

func (f Foo) Name() string {
	return f.name
}

func CustomConverter(a string) int {
	i, _ := strconv.Atoi(a)
	return i
}

//go:generate go run main.go -type Converter
type Converter interface {
	Convert(a Foo) (Bar, error) // Accepts err.
	ConvertImport(f foo.Foo) (b bar.Bar)
	ConvertImportPointer(f *foo.Foo) (b *bar.Bar)
	//ConvertFoo(a []Foo) ([]Bar, error) // Accepts err.
}

type Bar struct {
	Name string
	Age  int
	Task Task
	ID   uuid.UUID
}

func main() {
	if err := mapper.New(generateConverterFromFields); err != nil {
		log.Fatalln(err)
	}
}

func generateConverterFromFields(opt mapper.Option) error {
	var (
		pkgName  = opt.PkgName
		pkgPath  = opt.PkgPath
		typeName = opt.TypeName
		typ      = opt.Type
		out      = opt.Out
	)

	f := NewFile(pkgName) // e.g. main
	f.PackageComment(fmt.Sprintf("Code generated by %s, DO NOT EDIT.", Generator))

	spew.Dump(opt)
	generateConverter(f, typeName)
	generateConverterConstructor(f, typeName)
	for _, converter := range typ.InterfaceMethods {
		generateConvertMethod(f, pkgPath, typeName, converter)
	}

	return f.Save(out) // e.g. main_gen.go
}

func generateConverter(f *jen.File, typeName string) {
	// Output:
	//type Converter struct {
	//}

	f.Type().Id(typeName).Struct().Line()
}

func generateConverterConstructor(f *jen.File, typeName string) {
	// Output:
	// func NewConverter() *Converter {
	//   return &Converter{
	//   }
	// }

	f.Func().Id(fmt.Sprintf("New%s", typeName)).Params().Op("*").Id(typeName).Block(
		Return(Op("&").Id(typeName).Values(Dict{})),
	).Line()
}

func generateConvertMethod(f *jen.File, pkgPath, typeName string, fn mapper.FuncDto) {
	// Output:
	// func (c *Converter) Convert(a A) B {
	//   return B{
	//     Name: a.Name,
	//   }
	// }

	from, to := fn.From, fn.To

	dict := Dict{}
	for key, t := range to.Type.StructFields {
		f, ok := from.Type.StructFields[key]
		if ok {
			// Name: a.Name
			dict[Id(t.Name)] = Id(from.Name).Dot(f.Name)
			continue
		}
		var found bool
		for _, m := range from.Type.StructMethods {
			if m.Name == key && m.To.Type.Type == t.Type.Type {
				found = true
				// Name: a.Name()
				dict[Id(t.Name)] = Id(from.Name).Dot(m.Name).Call()
				break
			}
		}
		if !found {
			panic("method signature not found")
		}
	}

	var toReturnPtr, toPtr string
	if to.Type.IsPointer {
		toPtr = "&"
		toReturnPtr = "*"
	}

	var fromPtr string
	if from.Type.IsPointer {
		fromPtr = "*"
	}

	f.Func().Params(Id("c").Op("*").Id(typeName)). // (c *Converter)
							Id(fn.Name).Params(Id(from.Name).Op(fromPtr).Qual(skipImportIfBelongToSamePackage(pkgPath, from.Type.PkgPath), from.Type.Type)). // Convert(a A)
							Op(toReturnPtr).Qual(skipImportIfBelongToSamePackage(pkgPath, to.Type.PkgPath), to.Type.Type).Block(
		Return(Op(toPtr).Qual(skipImportIfBelongToSamePackage(pkgPath, to.Type.PkgPath), to.Type.Type).Values(dict)),
	).Line()
}

func skipImportIfBelongToSamePackage(pkgPath, fieldPkgPath string) string {
	if pkgPath == fieldPkgPath {
		return ""
	}
	return fieldPkgPath
}
