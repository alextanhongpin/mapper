// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

var _ Mapper = (*MapperImpl)(nil)

type MapperImpl struct {
	examplesURLBuilder *examples.URLBuilder
	uRLBuilder         *URLBuilder
}

func NewMapperImpl(examplesURLBuilder *examples.URLBuilder, uRLBuilder *URLBuilder) *MapperImpl {
	return &MapperImpl{
		examplesURLBuilder: examplesURLBuilder,
		uRLBuilder:         uRLBuilder,
	}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) (B, error) {
	a0URL, err := m.uRLBuilder.Build(a0.URL)
	if err != nil {
		return B{}, err
	}
	a0ExternalURL := m.examplesURLBuilder.Build(a0.ExternalURL)
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
