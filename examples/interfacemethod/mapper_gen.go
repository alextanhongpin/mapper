// Code gend by mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

type MapperImpl struct {
	examplesURLer examples.URLer
	uRLer         URLer
}

func NewMapperImpl(uRLer URLer, examplesURLer examples.URLer) *MapperImpl {
	return &MapperImpl{
		examplesURLer: examplesURLer,
		uRLer:         uRLer,
	}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) (B, error) {
	return B{
		ExternalURL: m.examplesURLer.Build(a0.ExternalURL),
		URL:         m.uRLer.Build(a0.URL),
	}, nil
}

func (m *MapperImpl) AtoB(a0 A) (B, error) {
	return m.mapMainAToMainB(a0)
}
