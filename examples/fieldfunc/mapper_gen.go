// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
package main

import (
	"database/sql"
	examples "github.com/alextanhongpin/mapper/examples"
	uuid "github.com/google/uuid"
	"strconv"
)

var _ Mapper = (*MapperImpl)(nil)

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) (B, error) {
	a0ID := IntToString(a0.ID)
	var a0IDs []uuid.UUID
	for _, each := range a0.IDs {
		tmp, err := uuid.Parse(each)
		if err != nil {
			return B{}, err
		}
		a0IDs = append(a0IDs, tmp)
	}
	a0ExternalID := examples.IntToString(a0.ExternalID)
	var a0Nums []int
	for _, each := range a0.Nums {
		tmp, err := strconv.Atoi(each)
		if err != nil {
			return B{}, err
		}
		a0Nums = append(a0Nums, tmp)
	}
	a0UUID, err := uuid.Parse(a0.UUID)
	if err != nil {
		return B{}, err
	}
	a0Remarks := NullStringToPointer(a0.Remarks)
	a0RemarksError, err := NullStringToPointerError(a0.RemarksError)
	if err != nil {
		return B{}, err
	}
	var a0PtrString sql.NullString
	if a0.PtrString != nil {
		a0PtrString = PointerStringToNullString(a0.PtrString)
	}
	return B{
		ExternalID:   a0ExternalID,
		ID:           a0ID,
		IDs:          a0IDs,
		Nums:         a0Nums,
		PtrString:    a0PtrString,
		Remarks:      a0Remarks,
		RemarksError: a0RemarksError,
		UUID:         a0UUID,
	}, nil
}

func (m *MapperImpl) mapExamplesCustomFieldToMainCustomField(c0 examples.CustomField) (CustomField, error) {
	c0Num, err := StringToInt(c0.Num)
	if err != nil {
		return CustomField{}, err
	}
	return CustomField{Num: c0Num}, nil
}

func (m *MapperImpl) mapMainCToMainD(c0 C) D {
	c0ID := IntToString(c0.ID)
	return D{ID: c0ID}
}

func (m *MapperImpl) AtoB(a0 A) (B, error) {
	a1, err := m.mapMainAToMainB(a0)
	if err != nil {
		return B{}, err
	}
	return a1, nil
}

func (m *MapperImpl) ConvertImportedFunc(c0 examples.CustomField) (CustomField, error) {
	c1, err := m.mapExamplesCustomFieldToMainCustomField(c0)
	if err != nil {
		return CustomField{}, err
	}
	return c1, nil
}

func (m *MapperImpl) ConvertImportedFuncPointer(c0 examples.CustomField) (*CustomField, error) {
	c1, err := m.mapExamplesCustomFieldToMainCustomField(c0)
	if err != nil {
		return nil, err
	}
	c2 := &c1
	return c2, nil
}

func (m *MapperImpl) CtoD(c0 C) D {
	c1 := m.mapMainCToMainD(c0)
	return c1
}

func (m *MapperImpl) SliceAtoB(a0 []A) ([]B, error) {
	var a1 []B
	for _, each := range a0 {
		tmp, err := m.mapMainAToMainB(each)
		if err != nil {
			return nil, err
		}
		a1 = append(a1, tmp)
	}
	return a1, nil
}

func (m *MapperImpl) SliceCtoD(c0 []C) []D {
	c1 := make([]D, len(c0))
	for i, each := range c0 {
		c1[i] = m.mapMainCToMainD(each)
	}
	return c1
}

func (m *MapperImpl) VariadicAtoB(a0 ...A) ([]B, error) {
	var a1 []B
	for _, each := range a0 {
		tmp, err := m.mapMainAToMainB(each)
		if err != nil {
			return nil, err
		}
		a1 = append(a1, tmp)
	}
	return a1, nil
}

func (m *MapperImpl) VariadicCtoD(c0 ...C) []D {
	c1 := make([]D, len(c0))
	for i, each := range c0 {
		c1[i] = m.mapMainCToMainD(each)
	}
	return c1
}
