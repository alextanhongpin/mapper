# mapper

Generate code that maps struct to struct.

Similar projects:
- [builder](https://github.com/alextanhongpin/builder): generate builder for private fields
- [getter](https://github.com/alextanhongpin/getter): generate getters for private struct fields, with inlining

## Design thoughts
- should not include pointer. Mapping requires one struct and returns one struct or error.
- should not allow context (?). Is there a need to pass context in mapping?
- Elem refers to the base type, so slice or pointer type User has Elem User

See the examples folder for results.

# Basic

## Field

The mapper should be able to map type A to type B, given that both types:

- ~both must have the same number of fields~ the target fields can be more than the source fields, and they may reference the same source fields. All target fields must have a corresponding mapping from the source fields, unless they are ignored.
- ~both must not be a pointer (it does not make sense for mapping pointer to non-pointer)~ sql.NullString can be converted to pointer string. As long as there are imported functions that allows the transformation, it should be valid
- both names must match, and could be altered through alias
- both types must match. If there is an imported function F, then the param P must match the source S and the result R must match the target T. F(P)->R is F(S)->T

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

## Error
## Nested
## Slice and Variadic

# Tags
## Renaming Field and Methods
## Ignore
## Func
## Interface and Struct

## TODO

- [ ] better error handling
- [ ] handle exported and private fields
