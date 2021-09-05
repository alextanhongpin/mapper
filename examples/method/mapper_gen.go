// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

var _ Mapper = (*MapperImpl)(nil)

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(g0 A) B {
	g0Status := g0.Status()
	g1Status := StatusToString(g0Status)
	return B{
		Bool:   g0.Bool(),
		ID:     g0.ID(),
		Map:    g0.Map(),
		Slice:  g0.Slice(),
		Status: g1Status,
		Str:    g0.Str(),
		Time:   g0.Time(),
	}
}

func (m *MapperImpl) mapMainCToMainD(g0 C) (D, error) {
	g0ID, err := g0.ID()
	if err != nil {
		return D{}, err
	}
	return D{ID: g0ID}, nil
}

func (m *MapperImpl) mapExamplesAToExamplesB(g0 examples.A) examples.B {
	return examples.B{
		Bool:  g0.Bool,
		ID:    g0.ID,
		Map:   g0.Map,
		Slice: g0.Slice,
		Str:   g0.Str,
	}
}

func (m *MapperImpl) AtoB(g0 A) B {
	g1 := m.mapMainAToMainB(g0)
	return g1
}

func (m *MapperImpl) CtoD(g0 C) (D, error) {
	g1, err := m.mapMainCToMainD(g0)
	if err != nil {
		return D{}, err
	}
	return g1, nil
}

func (m *MapperImpl) ExternalAtoB(g0 examples.A) examples.B {
	g1 := m.mapExamplesAToExamplesB(g0)
	return g1
}
