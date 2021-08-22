// Code gend by mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapExamplesUserToMainUser(u0 examples.User) (User, error) {
	return User{
		ID:   u0.ID,
		Name: u0.Name,
	}, nil
}

func (m *MapperImpl) mapExamplesBookToMainBook(b0 examples.Book) (Book, error) {
	return Book{
		ID:     b0.ID,
		Title:  b0.Title,
		UserID: b0.UserID,
	}, nil
}

func (m *MapperImpl) mapExamplesPriceToMainPrice(p0 examples.Price) Price {
	return Price{
		Amount:   p0.Amount,
		Currency: p0.Currency,
	}
}

func (m *MapperImpl) ConvertBook(b0 examples.Book) (Book, error) {
	return m.mapExamplesBookToMainBook(b0)
}

func (m *MapperImpl) ConvertPrice(p0 examples.Price) Price {
	return m.mapExamplesPriceToMainPrice(p0)
}

func (m *MapperImpl) ConvertUser(u0 examples.User) (User, error) {
	return m.mapExamplesUserToMainUser(u0)
}
