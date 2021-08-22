// Code gend by mapper, DO NOT EDIT.
package main

import (
	examples "github.com/alextanhongpin/mapper/examples"
	uuid "github.com/google/uuid"
	"strconv"
)

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) (B, error) {
	a0IDs := make([]uuid.UUID, len(a0.IDs))
	for i, each := range a0.IDs {
		var err error
		a0IDs[i], err = uuid.Parse(each)
		if err != nil {
			return B{}, err
		}
	}

	a0Nums := make([]int, len(a0.Nums))
	for i, each := range a0.Nums {
		var err error
		a0Nums[i], err = strconv.Atoi(each)
		if err != nil {
			return B{}, err
		}
	}

	a0UUID, err := uuid.Parse(a0.UUID)
	if err != nil {
		return B{}, err
	}

	return B{
		ExternalID: examples.IntToString(a0.ExternalID),
		ID:         IntToString(a0.ID),
		IDs:        a0IDs,
		Nums:       a0Nums,
		UUID:       a0UUID,
	}, nil
}

func (m *MapperImpl) VariadicAtoB(a0 ...A) ([]B, error) {
	res := make([]B, len(a0))
	for i, each := range a0 {
		var err error
		res[i], err = m.mapMainAToMainB(each)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (m *MapperImpl) AtoB(a0 A) (B, error) {
	return m.mapMainAToMainB(a0)
}

func (m *MapperImpl) SliceAtoB(a0 []A) ([]B, error) {
	res := make([]B, len(a0))
	for i, each := range a0 {
		var err error
		res[i], err = m.mapMainAToMainB(each)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
