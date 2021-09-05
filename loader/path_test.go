package loader_test

import (
	"strings"
	"testing"

	"github.com/alextanhongpin/getter/loader"
)

func TestFileNameFromTypeName(t *testing.T) {

	tests := []struct {
		scenario string
		input    string
		output   string
		typename string
		expected string
	}{
		{
			scenario: "no output path",
			input:    "/input/path/to/hello.go",
			output:   "",
			typename: "Foo",
			expected: "/input/path/to/foo_gen.go",
		},
		{
			scenario: "output is dir",
			input:    "/input/path/to/hello.go",
			output:   "/output/path/to/",
			typename: "Foo",
			expected: "/output/path/to/foo_gen.go",
		},
		{
			scenario: "output is file",
			input:    "/input/path/to/hello.go",
			output:   "/output/path/to/Foo.go",
			typename: "Foo",
			expected: "/output/path/to/Foo.go",
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			got := loader.FileNameFromTypeName(test.input, test.output, test.typename)
			if !strings.HasSuffix(got, test.expected) {
				t.Fatalf("expected %s, got %s", test.expected, got)
			}
		})
	}

}
