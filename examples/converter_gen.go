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

func (c *Converter) Convert(a Foo) (Bar, error) {
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

func (c *Converter) ConvertFoo(a Foo) (Bar, error) {
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

func (c *Converter) ConvertImport(f foo.Foo) (bar.Bar, error) {
	fID, err := f.ID()
	if err != nil {
		return bar.Bar{}, err
	}

	return bar.Bar{
		ID:   fID,
		Name: f.Name,
	}, nil
}

func (c *Converter) ConvertImportPointer(f *foo.Foo) (*bar.Bar, error) {
	fID, err := f.ID()
	if err != nil {
		return &bar.Bar{}, err
	}

	return &bar.Bar{
		ID:   fID,
		Name: f.Name,
	}, nil
}

func (c *Converter) ConvertNameless(f Foo) (Bar, error) {
	fID, err := f.ID()
	if err != nil {
		return Bar{}, err
	}

	fCustomID, err := ParseUUID(f.CustomID)
	if err != nil {
		return Bar{}, err
	}

	return Bar{
		ExternalID: fCustomID,
		ID:         fID,
		Name:       f.Name(),
		RealAge:    f.FakeAge,
		Task:       f.Task,
	}, nil
}
