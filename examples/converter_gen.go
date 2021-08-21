// Code generated by mapper, DO NOT EDIT.
package main

import (
	bar "github.com/alextanhongpin/mapper/examples/bar"
	foo "github.com/alextanhongpin/mapper/examples/foo"
	uuid "github.com/google/uuid"
)

type ConverterImpl struct {
	customStructConverter    *CustomStructConverter
	customInterfaceConverter CustomInterfaceConverter
}

func NewConverterImpl(customStructConverter *CustomStructConverter, customInterfaceConverter CustomInterfaceConverter) *ConverterImpl {
	return &ConverterImpl{
		customInterfaceConverter: customInterfaceConverter,
		customStructConverter:    customStructConverter,
	}
}

func (c *ConverterImpl) mapMainFooToMainBar(f0 Foo) (Bar, error) {
	f0ID, err := f0.ID()
	if err != nil {
		return Bar{}, err
	}

	f0CustomID, err := uuid.Parse(f0.CustomID)
	if err != nil {
		return Bar{}, err
	}

	return Bar{
		ExternalID: f0CustomID,
		ID:         f0ID,
		Name:       f0.Name(),
		RealAge:    f0.FakeAge,
		Task:       f0.Task,
	}, nil
}

func (c *ConverterImpl) mapMainAToMainB(a0 A) B {
	return B{Name: a0.Name}
}

func (c *ConverterImpl) mapFooFooToBarBar(f0 foo.Foo) (bar.Bar, error) {
	f0ID, err := f0.ID()
	if err != nil {
		return bar.Bar{}, err
	}

	return bar.Bar{
		ID:   f0ID,
		Name: f0.Name,
	}, nil
}

func (c *ConverterImpl) mapMainCToMainD(c0 C) D {
	return D{
		Age: c.customInterfaceConverter.ConvertToString(c0.Age),
		ID:  c.customStructConverter.ConvertToString(c0.ID),
	}
}

func (c *ConverterImpl) mapMainDToMainC(d0 D) (C, error) {
	d0Age, err := c.customInterfaceConverter.ConvertToInt(d0.Age)
	if err != nil {
		return C{}, err
	}

	d0ID, err := c.customStructConverter.ConvertToInt(d0.ID)
	if err != nil {
		return C{}, err
	}

	return C{
		Age: d0Age,
		ID:  d0ID,
	}, nil
}

func (c *ConverterImpl) ConvertNameless(f0 Foo) (Bar, error) {
	return c.mapMainFooToMainBar(f0)
}

func (c *ConverterImpl) ConvertSlice(a0 []Foo) ([]Bar, error) {
	var err error
	res := make([]Bar, len(a0))
	for i, s := range a0 {
		res[i], err = c.mapMainFooToMainBar(s)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (c *ConverterImpl) ConvertSliceWithoutErrors(a0 []A) []B {
	res := make([]B, len(a0))
	for i, s := range a0 {
		res[i] = c.mapMainAToMainB(s)
	}
	return res
}

func (c *ConverterImpl) Convert(a0 Foo) (Bar, error) {
	return c.mapMainFooToMainBar(a0)
}

func (c *ConverterImpl) ConvertImport(f0 foo.Foo) (bar.Bar, error) {
	return c.mapFooFooToBarBar(f0)
}

func (c *ConverterImpl) ConvertImportStruct(c0 C) D {
	return c.mapMainCToMainD(c0)
}

func (c *ConverterImpl) ConvertImportStructWithError(d0 D) (C, error) {
	return c.mapMainDToMainC(d0)
}
