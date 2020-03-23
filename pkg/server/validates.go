package server

import (
	"fmt"
	"strings"

	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/tags"
)

func validateInterface(iface types.Interface) error {

	var errs []error

	if len(iface.Methods) == 0 {
		errs = append(errs, fmt.Errorf("%s does not have any methods", iface.Name))
	}

	for _, m := range iface.Methods {
		errs = append(errs, validateFunction(m)...)
	}
	return composeErrors(errs...)
}

func validateFunction(fn *types.Function) (errs []error) {

	tags := tags.ParseTags(fn.Docs)

	// don't validate when `@gkg -` provided
	if tags.IsSet("disable") {
		return
	}

	if !isErrorLast(fn.Results) {
		errs = append(errs, fmt.Errorf("%s: last result should be of type error", fn.Name))
	}
	for _, param := range append(fn.Args, fn.Results...) {

		if param.Name == "" {
			errs = append(errs, fmt.Errorf("%s: unnamed parameter of type %s", fn.Name, param.Type.String()))
		}

		if iface := types.TypeInterface(param.Type); iface != nil && !iface.(types.TInterface).Interface.IsEmpty() {
			errs = append(errs, fmt.Errorf("%s: non empty interface %s is not allowed, delcare it outside", fn.Name, param.String()))
		}

		if strct := types.TypeStruct(param.Type); strct != nil {
			errs = append(errs, fmt.Errorf("%s: raw struct %s is not allowed, declare it outside", fn.Name, param.Name))
		}

		if f := types.TypeFunction(param.Type); f != nil {
			errs = append(errs, fmt.Errorf("%s: raw function %s is not allowed, declare it outside", fn.Name, param.Name))
		}
	}
	return
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

func composeErrors(errs ...error) error {

	if len(errs) > 0 {
		var strs []string

		for _, err := range errs {
			if err != nil {
				strs = append(strs, err.Error())
			}
		}

		if len(strs) == 1 {
			return fmt.Errorf(strs[0])
		}

		if len(strs) > 0 {
			return fmt.Errorf("many errors:\n%v", strings.Join(strs, "\n"))
		}
	}
	return nil
}
