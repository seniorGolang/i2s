package server

import (
	"context"
	"strconv"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/utils"
)

func removeContextIfFirst(fields []types.Variable) []types.Variable {
	if IsContextFirst(fields) {
		return fields[1:]
	}
	return fields
}

func removeSkippedFields(fields []types.Variable, skipFields []string) []types.Variable {

	var result []types.Variable

	for _, field := range fields {
		add := true
		for _, skip := range skipFields {
			if strings.TrimSpace(skip) == field.Name {
				add = false
				break
			}
		}
		if add {
			result = append(result, field)
		}
	}

	return result
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

func structField(ctx context.Context, field *types.Variable) *Statement {

	fieldName := field.Name
	s := Id(utils.ToCamel(fieldName))

	s.Add(fieldType(ctx, field.Type, false))
	s.Tag(map[string]string{"json": utils.ToLowerCamel(fieldName)})
	if types.IsEllipsis(field.Type) {
		s.Comment("This field was defined with ellipsis (...).")
	}
	return s
}

func fieldType(ctx context.Context, field types.Type, allowEllipsis bool) *Statement {

	c := &Statement{}

	imported := false

	for field != nil {
		switch f := field.(type) {
		case types.TImport:
			if f.Import != nil {

				if srcFile, ok := ctx.Value("code").(*File); ok {
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
				c.Index(Lit(f.ArrayLen))
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

func interfaceType(ctx context.Context, p *types.Interface) (code []Code) {
	for _, x := range p.Methods {
		code = append(code, functionDefinition(ctx, x))
	}
	return
}

func functionDefinition(ctx context.Context, signature *types.Function) *Statement {
	return Id(signature.Name).
		Params(funcDefinitionParams(ctx, signature.Args)).
		Params(funcDefinitionParams(ctx, signature.Results))
}

func funcDefinitionParams(ctx context.Context, fields []types.Variable) *Statement {
	c := &Statement{}
	c.ListFunc(func(g *Group) {
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

func methodDefinition(ctx context.Context, obj string, signature *types.Function) *Statement {
	return Func().
		Params(Id(obj).Id(obj)).
		Add(functionDefinition(ctx, signature))
}

func paramNames(fields []types.Variable) *Statement {
	var list []Code
	for _, field := range fields {
		v := Id(utils.ToLowerCamel(field.Name))
		if types.IsEllipsis(field.Type) {
			v.Op("...")
		}
		list = append(list, v)
	}
	return List(list...)
}

func nameOfLastResultError(fn *types.Function) string {
	if IsErrorLast(fn.Results) {
		return fn.Results[len(fn.Results)-1].Name
	}
	return "err"
}

func structFieldName(field *types.Variable) *Statement {
	return Id(utils.ToCamel(field.Name))
}

type FileProxy struct {
	*File
}

func NewFileProxy(packageName string) *FileProxy {
	return &FileProxy{
		File: NewFile(packageName),
	}
}

// Save method call Save of original File and call goimports on saved file
func (fw *FileProxy) Save(path string) (err error) {

	err = fw.File.Save(path)
	if err != nil {
		return
	}

	return utils.GoImports(path)
}
