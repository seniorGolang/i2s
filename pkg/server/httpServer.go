package server

import (
	"os"
	"path"
	"strings"

	. "github.com/dave/jennifer/jen"

	"github.com/seniorGolang/i2s/pkg/meta"
	"github.com/seniorGolang/i2s/pkg/tags"
	"github.com/seniorGolang/i2s/pkg/utils"
)

func renderHttpServer(info *meta.GenerationInfo) (err error) {

	srcFile := NewFileProxy("server")

	srcFile.PackageComment("GENERATED BY i2s. DO NOT EDIT.")

	srcFile.ImportName(packageTrace, "trace")
	srcFile.ImportName(packagePathHttp, "http")
	srcFile.ImportName(packageGorillaMux, "mux")
	srcFile.ImportName(packageKitServer, "server")
	srcFile.ImportName(packagePathPackageUtils, "utils")
	srcFile.ImportName(packageOpentracing, "opentracing")
	srcFile.ImportAlias(packagePathGoKitTransportHTTP, "kitHttp")

	for _, iface := range info.Services {
		service := strings.ToLower(iface.Name)
		srcFile.ImportName(path.Join(info.BasePackageImport, "transport", service), service)
	}

	srcFile.Add(appServeHttp(info))
	srcFile.Add(httpServerHandler(info))
	srcFile.Line().Add(caselessMatcher())
	srcFile.Line().Add(httpServer(info))

	filePath := path.Join(info.OutputFilePath, "transport", "server")

	if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
		return
	}
	return srcFile.Save(path.Join(filePath, "http.go"))
}

func appServeHttp(info *meta.GenerationInfo) (code *Statement) {

	return Func().Params(Id("app").Op("*").Id("appServer")).Id("ServeHTTP").Params(Id("opts").Op("...").Qual(packagePathGoKitTransportHTTP, "ServerOption")).Block(
		Id("app").Op(".").Id("srvHttp").Op("=").Qual(packageKitServer, "StartFastHttpServer").Call(

			Id("httpServerHandler").CallFunc(func(g *Group) {

				for _, iface := range info.Services {
					service := utils.ToLowerCamel(iface.Name)
					g.Id("app").Op(".").Id(service + "Endpoints")
				}

				g.Id("opts").Op("...")

			}), Id("app").Op(".").Id("config").Op(".").Id("BindAddr").Call()),
		Return(),
	)
}

func httpServerHandler(info *meta.GenerationInfo) Code {

	return Line().Func().Id("httpServerHandler").ParamsFunc(func(p *Group) {

		for _, iface := range info.Services {
			service := utils.ToLowerCamel(iface.Name)
			p.Id(service+"Endpoints").Qual(path.Join(info.BasePackageImport, "transport", strings.ToLower(service)), endpointsSetName)
		}

		p.Id("opts").Op("...").Qual(packagePathGoKitTransportHTTP, "ServerOption")

	}).Params(
		Qual(packagePathHttp, "Handler"),
	).BlockFunc(
		func(group *Group) {

			group.Line()

			group.Id("before").Op(":=").Qual(packagePathGoKitTransportHTTP, "ServerBefore").Call(
				Func().Params(Id(_ctx_).Qual(packagePathContext, "Context"), Id("r").Op("*").Qual(packagePathHttp, "Request")).Qual(packagePathContext, "Context").BlockFunc(func(g *Group) {

					g.Line()
					g.Id(_ctx_).Op("=").Qual(packagePathPackageUtils, "HttpToContext").Call(Id(_ctx_), Id("r"))
					g.Line()

					g.Id("span").Op(":=").Qual(packageTrace, "SpanFromHttp").Call(Qual(packagePathFmt, "Sprintf").Call(Lit("http:%s"), Id("r").Dot("URL").Dot("Path")), Id("r"))
					g.Id("span").Dot("SetTag").Call(Lit("requestPath"), Id("r").Dot("URL").Dot("Path"))
					g.Id(_ctx_).Op("=").Qual(packageOpentracing, "ContextWithSpan").Call(Id(_ctx_), Id("span"))
					g.Line()

					g.Return(Id(_ctx_))
				}),
			)

			group.Line()

			group.Id("after").Op(":=").Qual(packagePathGoKitTransportHTTP, "ServerAfter").Call(
				Func().Params(Id(_ctx_).Qual(packagePathContext, "Context"), Id("w").Qual(packagePathHttp, "ResponseWriter")).Qual(packagePathContext, "Context").BlockFunc(func(g *Group) {

					g.Line()
					g.If(Id("span").Op(":=").Qual(packageOpentracing, "SpanFromContext").Call(Id(_ctx_)).Op(";").Id("span").Op("!=").Nil()).Block(
						Id("span").Dot("SetTag").Call(Lit("responseCode"), Qual(packagePathHttp, "StatusOK")),
						Id("span").Dot("Finish").Call(),
					)

					g.Return(Id(_ctx_))
				}),
			)

			group.Line()

			group.Id("errorEncoder").Op(":=").Qual(packagePathGoKitTransportHTTP, "ServerErrorEncoder").Call(
				Func().Params(Id(_ctx_).Qual(packagePathContext, "Context"), Id("err").Error(), Id("w").Qual(packagePathHttp, "ResponseWriter")).BlockFunc(func(g *Group) {

					g.Line()
					g.Qual(packagePathGoKitTransportHTTP, "DefaultErrorEncoder").Call(Id(_ctx_), Id("err"), Id("w"))
					g.Id("span").Op(":=").Qual(packageOpentracing, "SpanFromContext").Call(Id(_ctx_))
					g.If(Id("span").Op("==").Nil()).Block(Return())

					g.Line()
					g.Id("span").Dot("SetTag").Call(Lit("err"), Id("err").Dot("Error").Call())
					g.Id("responseCode").Op(":=").Qual(packagePathHttp, "StatusInternalServerError")
					g.If(List(Id("sc"), Id("ok")).Op(":=").Id("err").Assert(Qual(packagePathGoKitTransportHTTP, "StatusCoder")).Op(";").Id("ok")).Block(
						Id("responseCode").Op("=").Id("sc").Dot("StatusCode").Call(),
					)
					g.Id("span").Dot("SetTag").Call(Lit("responseCode"), Id("responseCode"))

					g.Line()
					g.Id("span").Dot("Finish").Call()
				}),
			)

			group.Line()

			group.Id("opts").Op("=").Id("append").Call(
				Id("opts"),
				Line().Func().Params(Id("s").Op("*").Qual(packagePathGoKitTransportHTTP, "Server")).Block(
					Id("before").Call(Id("s")),
				),
				Func().Params(Id("s").Op("*").Qual(packagePathGoKitTransportHTTP, "Server")).Block(
					Id("after").Call(Id("s")),
				),
				Func().Params(Id("s").Op("*").Qual(packagePathGoKitTransportHTTP, "Server")).Block(
					Id("errorEncoder").Call(Id("s")),
				),
			)

			group.Line().Return(Id("newHTTPHandler").CallFunc(func(g *Group) {

				for _, iface := range info.Services {
					service := utils.ToLowerCamel(iface.Name)
					g.Id(service + "Endpoints")
				}
				g.Id("opts").Op("...")
			}))
		})
}

func caselessMatcher() *Statement {

	return Func().Id("caselessMatcher").Params(Id("next").Qual(packagePathHttp, "Handler")).Params(Qual(packagePathHttp, "Handler")).Block(
		Return(
			Qual(packagePathHttp, "HandlerFunc").Call(
				Func().Params(Id("w").Qual(packagePathHttp, "ResponseWriter"), Id("r").Op("*").Qual(packagePathHttp, "Request")).Block(
					Id("r").Dot("URL").Dot("Path").Op("=").Qual(packagePathStrings, "ToLower").Call(Id("r").Dot("URL").Dot("Path")),
					Id("next").Dot("ServeHTTP").Call(Id("w"), Id("r")),
				),
			),
		),
	)
}

func httpServer(info *meta.GenerationInfo) *Statement {

	return Func().Id("newHTTPHandler").ParamsFunc(func(p *Group) {

		for _, iface := range info.Services {
			service := utils.ToLowerCamel(iface.Name)
			p.Id(service+"Endpoints").Qual(path.Join(info.BasePackageImport, "transport", strings.ToLower(service)), endpointsSetName)
		}

		p.Id("opts").Op("...").Qual(packagePathGoKitTransportHTTP, "ServerOption")
	}).Params(
		Qual(packagePathHttp, "Handler"),
	).BlockFunc(func(g *Group) {

		g.Line().Id("mux").Op(":=").Qual(packageGorillaMux, "NewRouter").Call()

		for _, iface := range info.Services {

			service := utils.ToLowerCamel(iface.Name)
			serviceTags := tags.ParseTags(iface.Docs)

			if _, ok := serviceTags["jsonRPC-server"]; ok {

				var pathPrefix string
				if prefix, ok := serviceTags["jsonRPC-prefix"]; ok {
					pathPrefix = prefix
				}

				httpPath := path.Join(pathPrefix, strings.ToLower(service), "{method}")
				httpBasePath := path.Join(pathPrefix, strings.ToLower(service))

				g.Id("mux").Dot("Methods").Call(Lit("POST")).Dot("Path").
					Call(Lit("/" + httpBasePath)).Dot("Handler").Call(Id("jsonRpcServerHandler").Call(
					Qual(path.Join(info.BasePackageImport, "transport", strings.ToLower(service)), "EndpointsToMap").Call(
						Id(service + "Endpoints"),
					)))

				g.Id("mux").Dot("Methods").Call(Lit("POST")).Dot("Path").
					Call(Lit("/" + httpPath)).Dot("Handler").Call(Id("jsonRpcServerHandler").Call(
					Qual(path.Join(info.BasePackageImport, "transport", strings.ToLower(service)), "EndpointsToMap").Call(
						Id(service + "Endpoints"),
					))).Line()
			}

			if _, ok := serviceTags["http-server"]; ok {

				for _, fn := range iface.Methods {

					if tags.ParseTags(fn.Docs).IsSet("disable-http") {
						continue
					}

					encoderName := commonHTTPResponseEncoderName
					decodeName := utils.ToCamel(decodeRequestName(fn) + "Http")

					if name := tags.ParseTags(fn.Docs).Value("http-encoder", "-"); name != "-" {
						encoderName = name
					}

					if name := tags.ParseTags(fn.Docs).Value("http-decoder", "-"); name != "-" && name != "" {
						decodeName = name
					}

					httpPath := path.Join(tags.ParseTags(fn.Docs).Value("http-path", utils.ToLowerCamel(fn.Name)))
					method := tags.ParseTags(fn.Docs).Value("http-method", "POST")

					g.Id("mux").Dot("Methods").Call(Lit(method)).Dot("Path").
						Call(Lit("/" + httpPath)).Dot("Handler").Call(
						Line().Qual(packagePathGoKitTransportHTTP, "NewServer").Call(
							Line().Id(service+"Endpoints").Dot(endpointsStructFieldName(fn.Name)),
							Line().Qual(path.Join(info.BasePackageImport, "transport", strings.ToLower(service)), decodeName),
							Line().Qual(path.Join(info.BasePackageImport, "transport", strings.ToLower(service)), encoderName),
							Line().Id("opts").Op("...")),
					).Line()
				}
			}
		}
		g.Return(Id("accessControl").Call(Id("caselessMatcher").Call(Id("mux"))))
	})
}
