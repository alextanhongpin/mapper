package foo

import "github.com/google/uuid"

type Foo struct {
	id   string
	Name string
}

func (f Foo) ID() (uuid.UUID, error) {
	return uuid.Parse(f.id)
}
