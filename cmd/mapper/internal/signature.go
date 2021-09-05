package internal

import (
	"fmt"
	"go/types"
	"strings"
)

func PrettyFuncSignature(fn *types.Func) string {
	sig := fn.Type().(*types.Signature)
	rep := fmt.Sprintf("func %s", fn.Name())
	fnstr := types.TypeString(sig, (*types.Package).Name)
	fnstr = strings.ReplaceAll(fnstr, "func", rep)
	return fnstr
}
