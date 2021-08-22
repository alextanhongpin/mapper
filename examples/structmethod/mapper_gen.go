// Code gend by mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

type MapperImpl struct {
	uRLBuilder         *URLBuilder
	examplesURLBuilder *examples.URLBuilder
}

func NewMapperImpl(uRLBuilder *URLBuilder, examplesURLBuilder *examples.URLBuilder) *MapperImpl {
	return &MapperImpl{
		examplesURLBuilder: examplesURLBuilder,
		uRLBuilder:         uRLBuilder,
	}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) (B, error) {
	return B{
		ExternalURL: m.examplesURLBuilder.Build(a0.ExternalURL),
		URL:         m.uRLBuilder.Build(a0.URL),
	}, nil
}

func (m *MapperImpl) AtoB(a0 A) (B, error) {
	return m.mapMainAToMainB(a0)
}
