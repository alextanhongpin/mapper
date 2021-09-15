// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.

package main

import examples "github.com/alextanhongpin/mapper/examples"

var _ Mapper = (*MapperImpl)(nil)

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapExamplesBookToMainBook(b0 examples.Book) Book {
	b0Price := m.mapExamplesPriceToMainPrice(b0.Price)
	b1Price := &b0Price
	return Book{
		ID:     b0.ID,
		Price:  b1Price,
		Title:  b0.Title,
		UserID: b0.UserID,
	}
}

func (m *MapperImpl) mapExamplesPriceToMainPrice(p0 examples.Price) Price {
	return Price{
		Amount:   p0.Amount,
		Currency: &p0.Currency,
	}
}

func (m *MapperImpl) mapExamplesUserToMainUser(u0 examples.User) User {
	u0Books := make([]Book, len(u0.Books))
	for i, each := range u0.Books {
		u0Books[i] = m.mapExamplesBookToMainBook(each)
	}
	return User{
		Books: u0Books,
		ID:    u0.ID,
		Name:  u0.Name,
	}
}

func (m *MapperImpl) ConvertBook(b0 examples.Book) (Book, error) {
	b1 := m.mapExamplesBookToMainBook(b0)
	return b1, nil
}

func (m *MapperImpl) ConvertPrice(p0 examples.Price) *Price {
	p1 := m.mapExamplesPriceToMainPrice(p0)
	p2 := &p1
	return p2
}

func (m *MapperImpl) ConvertUser(u0 examples.User) (User, error) {
	u1 := m.mapExamplesUserToMainUser(u0)
	return u1, nil
}
