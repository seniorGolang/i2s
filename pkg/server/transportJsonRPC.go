package server

import (
	"context"
	"fmt"
	"path"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/meta"
	"github.com/seniorGolang/i2s/pkg/tags"
	"github.com/seniorGolang/i2s/pkg/utils"
)

func renderTransportJsonRPC(info *meta.GenerationInfo) (err error) {

	var requestDecoders []*types.Function
	var responseDecoders []*types.Function

	for _, fn := range info.Iface.Methods {
		requestDecoders = append(requestDecoders, fn)
		responseDecoders = append(responseDecoders, fn)
	}

	srcFile := NewFileProxy(info.PkgName)

	srcFile.PackageComment("GENERATED BY i2s. DO NOT EDIT.")

	ctx := context.WithValue(context.Background(), "code", srcFile)

	f := &Statement{}

	// f.Add(jsonRpcServerHandler(info))

	f.Line().Add(encodeResponse())

	f.Line().Add(mapServerEndPoints(info))

	for _, signature := range requestDecoders {
		f.Line().Add(decodeRequest(ctx, signature))
	}

	srcFile.ImportName(packagePathContext, "context")
	srcFile.ImportName(info.SourcePackageImport, serviceAlias)
	srcFile.ImportName(packagePathPackageJsonRPC, "jsonrpc")
	srcFile.Add(f)

	return srcFile.Save(path.Join(info.OutputFilePath, "transport", strings.ToLower(info.ServiceName), "jsonRPC.go"))
}

func mapServerEndPoints(info *meta.GenerationInfo) Code {

	return Line().Func().Id("EndpointsToMap").ParamsFunc(func(p *Group) { p.Id("endpoints").Id(endpointsSetName) }).
		Params(Qual(packagePathPackageJsonRPC, "EndpointCodecMap")).
		BlockFunc(
			func(group *Group) {

				methods := make(Dict)
				for _, fn := range info.Iface.Methods {
					methods[Lit(utils.ToLowerCamel(fn.Name))] = Qual(packagePathPackageJsonRPC, "EndpointCodec").Values(
						Dict{
							Id("Endpoint"): Id("endpoints").Op(".").Id(endpointsStructFieldName(utils.ToCamel(fn.Name))),
							Id("Decode"):   Id(decodeRequestName(fn)),
							Id("Encode"):   Id(encodeResponseFunc),
						})
				}
				group.Return(Qual(packagePathPackageJsonRPC, "EndpointCodecMap").Values(methods))
			})
}

func encodeRequest() Code {

	fullName := "request"

	return Line().Func().Id(encodeRequestFunc).Params(Op("_").Qual(packagePathContext, "Context"), Id(fullName).Interface()).
		Params(Qual(packagePathJson, "RawMessage"), Error()).BlockFunc(
		func(group *Group) {
			group.Return().Qual(packagePathJson, "Marshal").Call(Id(fullName))
		})
}

func encodeResponse() Code {

	fullName := "response"

	return Line().Func().Id(encodeResponseFunc).Params(Op("_").Qual(packagePathContext, "Context"), Id(fullName).Interface()).
		Params(Qual(packagePathJson, "RawMessage"), Error()).BlockFunc(
		func(group *Group) {
			group.Return().Qual(packagePathJson, "Marshal").Call(Id(fullName))
		})
}

func decodeRequest(ctx context.Context, fn *types.Function) Code {

	fullName := "request"
	shortName := "req"
	params := removeContextIfFirst(fn.Args)

	return Line().Func().Id(decodeRequestName(fn)).Params(Id(_ctx_).Qual(packagePathContext, "Context"), Id(fullName).Qual(packagePathJson, "RawMessage")).
		Params(Id("result").Interface(), Id("err").Error()).BlockFunc(func(group *Group) {

		type headerArg struct {
			name string
			arg  types.Variable
		}
		headerArgs := make(map[string]*headerArg)

		if httpHeaders := tags.ParseTags(fn.Docs).Value("http-headers", ""); httpHeaders != "" {

			headerPairs := strings.Split(httpHeaders, ",")

			for _, pair := range headerPairs {
				if pairTokens := strings.Split(pair, "|"); len(pairTokens) == 2 {
					arg := strings.TrimSpace(pairTokens[0])
					header := strings.ToLower(strings.TrimSpace(pairTokens[1]))
					headerArgs[arg] = &headerArg{name: header}
				}
			}
		}

		var args []types.Variable
		for _, arg := range params {
			if _, found := headerArgs[arg.Name]; !found {
				args = append(args, arg)
			} else {
				headerArgs[arg.Name].arg = arg
			}
		}

		for k, v := range headerArgs {
			if v.arg.Base.Name == "" {
				delete(headerArgs, k)
			}
		}

		group.Line().Var().Id(shortName).Id(requestStructName(fn))

		group.If(Err().Op("=").Qual(packagePathJson, "Unmarshal").Call(Id(fullName), Op("&").Id(shortName)).Op(";").Err().Op("!=").Nil()).Block(
			Return(),
		)

		if len(headerArgs) > 0 {
			group.Line()
		}

		for argName, header := range headerArgs {

			toName := fmt.Sprintf("%s.%s", shortName, utils.ToCamel(argName))

			headerArgName := "_" + header.arg.Name

			group.If(List(Id(headerArgName), Id("ok")).Op(":=").Id(_ctx_).Dot("Value").Call(Lit(header.name)).Op(".").Call(String()).Op(";").Id("ok")).Block(
				stringToTypeConverter(header.arg, toName, Id(headerArgName), false),
			)
		}
		group.Line().Id("result").Op("=").Id(shortName)
		group.Return()
	})
}

func decodeRequestName(f *types.Function) string {
	return "decode" + f.Name + "Request"
}

func decodeResponseName(f *types.Function) string {
	return "decode" + f.Name + "Response"
}

func encodeResponseName(f *types.Function) string {
	return "encode" + f.Name + "Response"
}
