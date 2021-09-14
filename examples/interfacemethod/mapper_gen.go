// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

var _ Mapper = (*MapperImpl)(nil)

type MapperImpl struct {
	examplesURLer examples.URLer
	uRLer         URLer
}

func NewMapperImpl(examplesURLer examples.URLer, uRLer URLer) *MapperImpl {
	return &MapperImpl{
		examplesURLer: examplesURLer,
		uRLer:         uRLer,
	}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) (B, error) {

	a0ExternalURL := m.examplesURLer.Build(a0.ExternalURL)

	a0URL, err := m.uRLer.Build(a0.URL)
	if err != nil {
		return B{}, err
	}
	return B{
		ExternalURL: a0ExternalURL,
		URL:         a0URL,
	}, nil
}

func (m *MapperImpl) AtoB(a0 A) (B, error) {

	a1, err := m.mapMainAToMainB(a0)
	if err != nil {
		return B{}, err
	}
	return a1, nil
}
