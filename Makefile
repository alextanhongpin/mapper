# https://pkg.go.dev/cmd/gofmt
# https://github.com/uber-go/gopatch
# $ go install github.com/uber-go/gopatch@latest
# Use gopatch for refactoring.
patch:
	gopatch -p error.patch ./...
	go vet ./...
