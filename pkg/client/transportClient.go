package client

import (
	"context"
	"path"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/meta"
	"github.com/seniorGolang/i2s/pkg/utils"
)

func renderTransportClient(info *meta.GenerationInfo) (err error) {

	srcFile := NewFile(strings.ToLower(info.ServiceName))
	srcFile.PackageComment("GENERATED BY i2s. DO NOT EDIT.")

	ctx, _ := prepareContext(info.SourceFilePath, info.Iface)
	ctx = context.WithValue(ctx, "code", srcFile)

	for _, signature := range info.Iface.Methods {
		srcFile.Add(serviceEndpointMethod(ctx, signature)).Line().Line()
	}
	return srcFile.Save(path.Join(info.OutputFilePath, strings.ToLower(info.ServiceName), "client.go"))
}

func serviceEndpointMethod(ctx context.Context, signature *types.Function) *Statement {

	return methodDefinitionFull(ctx, endpointsSetName, signature).
		BlockFunc(serviceEndpointMethodBody(signature))
}

func methodDefinitionFull(ctx context.Context, obj string, signature *types.Function) *Statement {
	return Func().
		Params(Id(_service_).Id(obj)).
		Add(functionDefinition(ctx, signature))
}

func serviceEndpointMethodBody(fn *types.Function) func(g *Group) {

	return func(g *Group) {

		args := removeContextIfFirst(fn.Args)
		results := removeErrorIfLast(fn.Results)

		var leftStatement, rightStatement *Statement

		if len(args) == 0 {
			rightStatement = Id(_service_).Dot(endpointsStructFieldName(fn.Name)).Call(Id(_ctx_).Op(",").Id("nil"))

		} else {
			g.Line().Id("req").Op(":=").Id(requestStructName(fn)).Values(utils.DictByNormalVariables(args, args))
			rightStatement = Id(_service_).Dot(endpointsStructFieldName(fn.Name)).Call(Id(_ctx_).Op(",").Id("req"))
		}

		if len(results) > 0 {
			g.Line().Add(Var().Id("resp").Id("interface{}"))
			leftStatement = Line().Id("resp").Op(",").Id("err").Op("=")
		} else {
			leftStatement = Line().Id("_").Op(",").Id("err").Op("=")
		}

		if leftStatement == nil {
			leftStatement = Op("_").Op(",").Id("err").Op("=")
		}

		if leftStatement != nil && rightStatement != nil {
			g.Add(leftStatement.Add(rightStatement))
		}

		if len(results) >= 1 {
			g.Line().If(Err().Op("!=").Nil().Block(Return())).Line()
		}

		if len(results) > 0 {
			g.Id("response").Op(":=").Id("resp").Assert(Id(responseStructName(fn)))
		}

		if len(results) > 0 {
			g.ReturnFunc(func(group *Group) {
				for _, field := range removeErrorIfLast(fn.Results) {
					group.Id("response").Op(".").Add(structFieldName(&field))
				}
				group.Id(nameOfLastResultError(fn))
			})
		} else {
			g.Return()
		}
	}
}
