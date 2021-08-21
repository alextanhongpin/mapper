# mapper


Generate code that maps struct to struct.

## Design thoughts
- should not include pointer. Mapping requires one struct and returns one struct or error.
- should not allow context (?). Is there a need to pass context in mapping?
