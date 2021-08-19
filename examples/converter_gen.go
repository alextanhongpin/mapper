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

func (c *Converter) Convert(a Foo) Bar {
	return Bar{
		Age:  a.Age,
		ID:   a.ID,
		Name: a.Name(),
		Task: a.Task,
	}
}

func (c *Converter) ConvertImport(f foo.Foo) bar.Bar {
	return bar.Bar{
		ID:   f.ID,
		Name: f.Name,
	}
}

func (c *Converter) ConvertImportPointer(f *foo.Foo) *bar.Bar {
	return &bar.Bar{
		ID:   f.ID,
		Name: f.Name,
	}
}
