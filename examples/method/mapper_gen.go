// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) B {
	a0Status := a0.Status()
	a1Status := StatusToString(a0Status)
	return B{
		Bool:   a0.Bool(),
		ID:     a0.ID(),
		Map:    a0.Map(),
		Slice:  a0.Slice(),
		Status: a1Status,
		Str:    a0.Str(),
		Time:   a0.Time(),
	}
}

func (m *MapperImpl) mapMainCToMainD(c0 C) (D, error) {
	c0ID, err := c0.ID()
	if err != nil {
		return D{}, err
	}
	return D{ID: c0ID}, nil
}

func (m *MapperImpl) mapExamplesAToExamplesB(a0 examples.A) examples.B {
	return examples.B{
		Bool:  a0.Bool,
		ID:    a0.ID,
		Map:   a0.Map,
		Slice: a0.Slice,
		Str:   a0.Str,
	}
}

func (m *MapperImpl) AtoB(a0A A) B {
	a1A := m.mapMainAToMainB(a0A)
	return a1A
}

func (m *MapperImpl) CtoD(c0C C) (D, error) {
	c1C, err := m.mapMainCToMainD(c0C)
	if err != nil {
		return D{}, err
	}
	return c1C, nil
}

func (m *MapperImpl) ExternalAtoB(a0A examples.A) examples.B {
	a1A := m.mapExamplesAToExamplesB(a0A)
	return a1A
}
