package mapper

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var tagRe *regexp.Regexp

func init() {
	var err error
	tagRe, err = regexp.Compile(`map:"(([\w.]+)?,?([\w.\/]+)?)"`)
	if err != nil {
		panic(fmt.Sprintf("mapper: compile tag regex error: %s", err))
	}
}

func NewTag(tag string) (*Tag, bool) {
	if !tagRe.MatchString(tag) {
		return nil, false
	}
	matches := tagRe.FindAllStringSubmatch(tag, -1)
	match := matches[0]

	matched := match[1]
	if matched == "" {
		return nil, false
	}

	name := match[2]
	pkgPath, expr := path.Split(match[3])
	pkgPath = stripSlash(pkgPath) // Removes trailing slash

	var typeName, fn string
	parts := strings.Split(expr, ".")

	switch len(parts) {
	case 1:
		fn = parts[0]
	case 2:
		typeName, fn = parts[0], parts[1]
	default:
		panic(fmt.Sprintf(`mapper: invalid tag %q`, tag))
	}

	return &Tag{
		Name:     name,
		PkgPath:  pkgPath,
		Pkg:      filepath.Base(pkgPath),
		TypeName: typeName,
		Func:     fn,
	}, true
}

type Tag struct {
	Name string
	// If the pkgPath is not set, we assume it to be the same as the current root directory.
	PkgPath string `example:"github.com/your.org/yourpkg"`
	Pkg     string `example:"yourpkg"`
	// If the `pkg` is not empty, it could most likely be an interface or struct method.
	TypeName string `example:"YourStruct|YourInterface"`
	// If `pkg` is empty, then this is a pure function import.
	Func string `example:"YourMethod"`
}

func (t Tag) HasFunc() bool {
	return t.Func != ""
}

// IsAlias returns true if there is a name suggested for
// mapping.
func (t Tag) IsAlias() bool {
	return t.Name != ""
}

func (t Tag) IsFunc() bool {
	return t.HasFunc() && t.TypeName == ""
}

func (t Tag) IsMethod() bool {
	return t.HasFunc() && t.TypeName != ""
}

func (t Tag) IsImported() bool {
	return t.PkgPath != ""
}
