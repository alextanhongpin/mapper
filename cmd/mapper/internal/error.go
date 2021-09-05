package internal

import (
	"errors"
	"fmt"
	"strings"
)

func PrettyError(msg string, args ...interface{}) error {
	msg = strings.TrimSpace(fmt.Sprintf(msg, args...))
	rows := strings.Split(msg, "\n")
	res := make([]string, len(rows))

	for i, row := range rows {
		res[i] = strings.TrimSpace(row)
	}

	return errors.New(strings.Join(res, "\n"))
}
