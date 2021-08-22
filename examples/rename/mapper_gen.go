// Code gend by mapper, DO NOT EDIT.
package main

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) B {
	return B{
		AnotherID: a0.ID,
		Name:      a0.MyName,
		ToB:       a0.FromA,
	}
}

func (m *MapperImpl) AtoB(a0 A) B {
	return m.mapMainAToMainB(a0)
}
