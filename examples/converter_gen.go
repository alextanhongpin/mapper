// Code generated by mapper, DO NOT EDIT.
package main

import (
	bar "github.com/alextanhongpin/mapper/examples/bar"
	foo "github.com/alextanhongpin/mapper/examples/foo"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) mapMainAToMainB(a A) B {
	return B{Name: a.Name}
}

func (c *Converter) mapMainFooToMainBar(a Foo) (Bar, error) {
	aID, err := a.ID()
	if err != nil {
		return Bar{}, err
	}

	aCustomID, err := ParseUUID(a.CustomID)
	if err != nil {
		return Bar{}, err
	}

	return Bar{
		ExternalID: aCustomID,
		ID:         aID,
		Name:       a.Name(),
		RealAge:    a.FakeAge,
		Task:       a.Task,
	}, nil
}

func (c *Converter) mapFooFooToBarBar(f foo.Foo) (bar.Bar, error) {
	fID, err := f.ID()
	if err != nil {
		return bar.Bar{}, err
	}

	return bar.Bar{
		ID:   fID,
		Name: f.Name,
	}, nil
}

func (c *Converter) ConvertSliceWithoutErrors(a []A) []B {
	res := make([]B, len(a))
	for i, s := range a {
		res[i] = c.mapMainAToMainB(s)
	}
	return res
}

func (c *Converter) Convert(a Foo) (Bar, error) {
	return c.mapMainFooToMainBar(a)
}

func (c *Converter) ConvertImport(f foo.Foo) (bar.Bar, error) {
	return c.mapFooFooToBarBar(f)
}

func (c *Converter) ConvertNameless(f Foo) (Bar, error) {
	return c.mapMainFooToMainBar(f)
}

func (c *Converter) ConvertSlice(a []Foo) ([]Bar, error) {
	var err error
	res := make([]Bar, len(a))
	for i, s := range a {
		res[i], err = c.mapMainFooToMainBar(s)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
