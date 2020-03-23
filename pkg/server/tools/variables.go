package tools

import (
	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/utils"
)

func DictByNormalVariables(fields []types.Variable, normals []types.Variable) Dict {

	if len(fields) != len(normals) {
		panic("len of fields and normals not the same")
	}

	return DictFunc(func(d Dict) {
		for i, field := range fields {
			d[structFieldName(&field)] = Id(utils.ToLowerCamel(normals[i].Name))
		}
	})
}

func structFieldName(field *types.Variable) *Statement {
	return Id(utils.ToCamel(field.Name))
}
