// rename demonstrates how to use the `map` tag to customize the name for the
// mapper.
package main

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	AtoB(A) B
}

type A struct {
	ID     int
	MyName string `map:"Name"` // Maps this field to `B.Name`.
}

type B struct {
	AnotherID int `json:"anotherId" map:"ID"` // Maps this field from `A.ID`.
	Name      string
}
