# mapper

Generate code that maps struct to struct.

## Design thoughts
- should not include pointer. Mapping requires one struct and returns one struct or error.
- should not allow context (?). Is there a need to pass context in mapping?


## Basic

```go
type Mapper interface {
	AtoB(A) B
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

## TODO

- [ ] Add ignore tags
- [ ] Name resolution when not unique
