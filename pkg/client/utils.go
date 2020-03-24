package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/utils"
)

func removeContextIfFirst(fields []types.Variable) []types.Variable {
	if IsContextFirst(fields) {
		return fields[1:]
	}
	return fields
}

func IsContextFirst(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[0].Type)
	return name != nil &&
		types.TypeImport(fields[0].Type) != nil &&
		types.TypeImport(fields[0].Type).Package == packagePathContext && *name == "Context"
}

func requestStructName(signature *types.Function) string {
	return utils.ToCamel(signature.Name + "Request")
}

func responseStructName(signature *types.Function) string {
	return utils.ToCamel(signature.Name + "Response")
}

func removeErrorIfLast(fields []types.Variable) []types.Variable {
	if IsErrorLast(fields) {
		return fields[:len(fields)-1]
	}
	return fields
}

func IsErrorLast(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[len(fields)-1].Type)
	return name != nil &&
		types.TypeImport(fields[len(fields)-1].Type) == nil &&
		*name == "error"
}

func structRequestFieldName(field types.Variable) string {
	return field.Name
}

func structField(ctx context.Context, field *types.Variable) *jen.Statement {

	fieldName := field.Name
	s := jen.Id(utils.ToCamel(fieldName))

	s.Add(fieldType(ctx, field.Type, false))
	s.Tag(map[string]string{"json": utils.ToLowerCamel(fieldName)})
	if types.IsEllipsis(field.Type) {
		s.Comment("This field was defined with ellipsis (...).")
	}
	return s
}

func fieldType(ctx context.Context, field types.Type, allowEllipsis bool) *jen.Statement {

	c := &jen.Statement{}

	imported := false

	for field != nil {
		switch f := field.(type) {
		case types.TImport:
			if f.Import != nil {

				if srcFile, ok := ctx.Value("code").(*jen.File); ok {
					srcFile.ImportName(f.Import.Package, f.Import.Base.Name)
					c.Qual(f.Import.Package, "")
				} else {
					c.Qual(f.Import.Package, "")
				}
				imported = true
			}
			field = f.Next
		case types.TName:
			if !imported && !types.IsBuiltin(f) {
				c.Qual(SourcePackageImport(ctx), f.TypeName)
			} else {
				c.Id(f.TypeName)
			}
			field = nil
		case types.TArray:
			if f.IsSlice {
				c.Index()
			} else if f.ArrayLen > 0 {
				c.Index(jen.Lit(f.ArrayLen))
			}
			field = f.Next
		case types.TMap:
			return c.Map(fieldType(ctx, f.Key, false)).Add(fieldType(ctx, f.Value, false))
		case types.TPointer:
			// c.Op(tools.Repeat("*", f.NumberOfPointers))
			c.Op("*")
			field = f.Next
		case types.TInterface:
			mhds := interfaceType(ctx, f.Interface)
			return c.Interface(mhds...)
		case types.TEllipsis:
			if allowEllipsis {
				c.Op("...")
			} else {
				c.Index()
			}
			field = f.Next
		default:
			return c
		}
	}
	return c
}

func interfaceType(ctx context.Context, p *types.Interface) (code []jen.Code) {
	for _, x := range p.Methods {
		code = append(code, functionDefinition(ctx, x))
	}
	return
}

func functionDefinition(ctx context.Context, signature *types.Function) *jen.Statement {
	return jen.Id(signature.Name).
		Params(funcDefinitionParams(ctx, signature.Args)).
		Params(funcDefinitionParams(ctx, signature.Results))
}

func funcDefinitionParams(ctx context.Context, fields []types.Variable) *jen.Statement {
	c := &jen.Statement{}
	c.ListFunc(func(g *jen.Group) {
		for _, field := range fields {
			g.Id(utils.ToLowerCamel(field.Name)).Add(fieldType(ctx, field.Type, true))
		}
	})
	return c
}

type normalizedFunction struct {
	types.Function
	parent *types.Function
}

func normalizeFunction(signature *types.Function) *normalizedFunction {
	newFunc := &normalizedFunction{parent: signature}
	newFunc.Name = signature.Name
	newFunc.Args = normalizeVariables(signature.Args, normalArgPrefix)
	newFunc.Results = normalizeVariables(signature.Results, normalResultPrefix)
	return newFunc
}

func normalizeVariables(old []types.Variable, prefix string) (new []types.Variable) {
	for i := range old {
		v := old[i]
		v.Name = prefix + strconv.Itoa(i)
		new = append(new, v)
	}
	return
}

func methodDefinition(ctx context.Context, obj string, signature *types.Function) *jen.Statement {
	return jen.Func().
		Params(jen.Id(obj).Id(obj)).
		Add(functionDefinition(ctx, signature))
}

func paramNames(fields []types.Variable) *jen.Statement {
	var list []jen.Code
	for _, field := range fields {
		v := jen.Id(utils.ToLowerCamel(field.Name))
		if types.IsEllipsis(field.Type) {
			v.Op("...")
		}
		list = append(list, v)
	}
	return jen.List(list...)
}

func nameOfLastResultError(fn *types.Function) string {
	if IsErrorLast(fn.Results) {
		return fn.Results[len(fn.Results)-1].Name
	}
	return "err"
}

func structFieldName(field *types.Variable) *jen.Statement {
	return jen.Id(utils.ToCamel(field.Name))
}

func decodeMethodName(method string) string {
	return fmt.Sprintf("decode%sResponse", utils.ToCamel(method))
}

func decodeMethoJsonRpcName(method string) string {
	return fmt.Sprintf("decode%sJsonRpcResponse", utils.ToCamel(method))
}

func decodeResponseName(f *types.Function) string {
	return "decode" + f.Name + "Response"
}
