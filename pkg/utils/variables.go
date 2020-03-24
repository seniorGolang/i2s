package utils

import (
	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"
)

func DictByNormalVariables(fields []types.Variable, normals []types.Variable) Dict {

	if len(fields) != len(normals) {
		panic("len of fields and normals not the same")
	}

	return DictFunc(func(d Dict) {
		for i, field := range fields {
			d[structFieldName(&field)] = Id(ToLowerCamel(normals[i].Name))
		}
	})
}

func structFieldName(field *types.Variable) *Statement {
	return Id(ToCamel(field.Name))
}
