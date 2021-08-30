// rename demonstrates how to use the `map` tag to customize the name for the
// mapper.
package main

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	AtoB(A) B
}

type A struct {
	ID    int
	Name  string
	FromA string
}

type B struct {
	AnotherID int `json:"anotherId" map:"ID"` // Maps this field from `A.ID`.
	Name      string
	ToB       string `map:"FromA"` // Maps this field from `A.FromA`.
}
