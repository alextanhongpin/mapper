// rename demonstrates how to use the `map` tag to customize the name for the
// mapper.
package main

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	AtoB(A) B
}

type A struct {
	Status string
}

func (a A) CustomStatus() string {
	return a.Status
}

type B struct {
	Status string `map:"CustomStatus()"` // Maps this field from `A.CustomStatus()`
}
