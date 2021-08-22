// Code gend by mapper, DO NOT EDIT.
package main

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) (B, error) {
	return B{
		ExternalURL: a0.ExternalURL,
		URL:         a0.URL,
	}, nil
}

func (m *MapperImpl) AtoB(a0 A) (B, error) {
	return m.mapMainAToMainB(a0)
}
