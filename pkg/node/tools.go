package node

import (
	"github.com/vetcher/go-astra/types"
)

const (
	packagePathContext = "context"
)

func removeContextIfFirst(fields []types.Variable) ([]types.Variable, bool) {
	if isContextFirst(fields) {
		return fields[1:], true
	}
	return fields, false
}

func removeErrorIfLast(fields []types.Variable) ([]types.Variable, bool) {
	if isErrorLast(fields) {
		return fields[:len(fields)-1], true
	}
	return fields, false
}

func isContextFirst(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[0].Type)
	return name != nil &&
		types.TypeImport(fields[0].Type) != nil &&
		types.TypeImport(fields[0].Type).Package == packagePathContext && *name == "Context"
}

func isErrorLast(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[len(fields)-1].Type)
	return name != nil &&
		types.TypeImport(fields[len(fields)-1].Type) == nil &&
		*name == "error"
}
