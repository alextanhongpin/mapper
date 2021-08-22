# mapper

Generate code that maps struct to struct.

## Design thoughts
- should not include pointer. Mapping requires one struct and returns one struct or error.
- should not allow context (?). Is there a need to pass context in mapping?


## Basic

The mapper should be able to map type A to type B, given that both types:

- both must have the same number of fields
- both must not be a pointer (it does not make sense for mapping pointer to non-pointer)
- both names must match
- both types must match (slice to slice)

```go
type Mapper interface {
	AtoB (A) B
}

type A struct {
	ID string
}

type B struct {
	ID string
}
```

Expected:

```go
type MapperImpl struct {}

func (m MapperImpl) AtoB(a A) B {
	return B{
		ID: a.ID
	}
}
```

## Method


When A has a method that maps to the name of B's field, we should be able to use that.

- if that method returns a tuple error, `AtoB` should also return a tuple error
- both names must match
- both types must match

```go
type Mapper interface {
	AtoB (A) B
}

type A struct {
	id string
}
func (a A) ID() string {
	return a.id
}

type B struct {
	ID string
}
```

Expected:
```go
type MapperImpl struct {}

func (m MapperImpl) AtoB(a A) B {
	return B{
		ID: a.ID()
	}
}
```

## TODO

- [ ] Add ignore tags
- [ ] Add variadic (?) support
- [ ] Name resolution when not unique
- [ ] replace all pointer receiver name `c` with `ShortName(name)`
