package main

import (
	examples "github.com/alextanhongpin/mapper/examples"
)

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	// Supports error as second return parameter.
	ConvertUser(examples.User) (User, error) // Fails when no error is set
	ConvertBook(examples.Book) (Book, error)
	ConvertPrice(examples.Price) Price
}

type User struct {
	ID    int
	Name  string
	Books []Book `map:",Mapper.ConvertBook"`
}

type Book struct {
	ID     int
	UserID int
	Title  string
	Price  Price `map:",Mapper.ConvertPrice"`
}

type Price struct {
	Currency string
	Amount   float64
}
