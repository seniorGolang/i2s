package skeleton

import (
	"fmt"
	"os"
	"path"

	. "github.com/dave/jennifer/jen"

	"github.com/seniorGolang/i2s/pkg/utils"
)

var fields = []string{"serviceName", "nodeName", "version", "gitSHA", "buildStamp", "buildNumber"}

func genConfig(meta metaInfo) (err error) {

	log.Info("generate config")

	configPath := path.Join(meta.baseDir, "pkg", "config")

	if err = os.MkdirAll(configPath, os.ModePerm); err != nil {
		return
	}

	if err = renderConfig(meta, configPath); err != nil {
		return
	}

	if err = renderConfigApp(meta, configPath); err != nil {
		return
	}
	return
}

func renderConfig(meta metaInfo, configPath string) (err error) {

	srcFile := NewFile("config")

	srcFile.ImportName(pkgOS, "os")
	srcFile.ImportName(pkgENV, "env")
	srcFile.ImportName(pkgLog, "log")
	srcFile.ImportName(pkgUtils, "utils")
	srcFile.ImportName(pkgReflect, "reflect")

	srcFile.Add(renderConfigStruct(meta))
	srcFile.Add(renderBuildInfo())
	srcFile.Add(renderInternalMethods())
	srcFile.Add(renderPrintConfig())

	return srcFile.Save(path.Join(configPath, "config.go"))
}

func renderPrintConfig() Code {

	return Func().Id("printConfig").Params().Block(
		Line(),
		Id("configType").Op(":=").Qual(pkgReflect, "ValueOf").Call(Id("Values").Call()),
		Line(),
		For(Id("i").Op(":=").Lit(0).Op(";").Id("i").Op("<").Id("configType").Dot("NumField").Call().Op(";").Id("i").Op("++")).Block(
			Line(),
			Id("field").Op(":=").Id("configType").Dot("Field").Call(Id("i")),
			Id("envKey").Op(":=").Id("configType").Dot("Type").Call().Dot("Field").Call(Id("i")).Dot("Tag").Dot("Get").Call(Lit("env")),
			Line(),
			If(Id("envKey").Op("!=").Lit("")).Block(
				List(Id("_"), Id("found")).Op(":=").Qual(pkgOS, "LookupEnv").Call(Id("envKey")),
				Qual(pkgLog, "Info").Call(Id("envKey"), Id("field").Dot("Interface").Call(), Lit("default"), Op("!").Id("found")),
			),
		),
	)
}

func renderBuildInfo() Code {

	var params []Code

	for _, field := range fields {
		if field != "nodeName" {
			params = append(params, Id(field))
		}
	}

	return Func().Id("SetBuildInfo").Params(List(params...).String()).BlockFunc(func(g *Group) {

		g.Line()

		g.Qual(pkgLog, "SetLevel").Call(Id("Values").Call().Dot("LogLevel"))
		g.Qual(pkgLog, "SetServiceName").Call(Id("serviceName"))

		g.Line()
		g.List(Id("nodeName"), Id("_")).Op(":=").Qual(pkgOS, "Hostname").Call()
		g.Line()

		g.Id("conf").Op(":=").Id("internalConfig").Call()

		for _, field := range fields {
			g.Id("setLinkedString").Call(Op("&").Id("conf").Dot(field), Id(field))
		}

		g.Line()

		g.Qual(pkgLog, "Info").CallFunc(func(c *Group) {
			for _, field := range fields {
				c.Lit(field)
				c.Id(utils.ToCamel(field)).Call()
			}
		})
		g.Id("printConfig").Call()
	})
}

func renderInternalMethods() (code *Statement) {

	code = &Statement{}

	code.Var().Id("configuration").Op("*").Id("config")

	for _, field := range fields {

		code.Line().Func().Id(utils.ToCamel(field)).Params().Params(String()).Block(
			Return(Id("getLinkedString").Call(Id("internalConfig").Call().Dot(field))),
		).Line()
	}

	code.Line().Func().Id("internalConfig").Params().Params(Op("*").Id("config")).Block(

		Line().If(Id("configuration").Op("==").Nil()).Block(
			Id("configuration").Op("=").Op("&").Id("config").Op("{}"),
			Line().If(Err().Op(":=").Qual(pkgENV, "Parse").Call(Id("configuration")).Op(";").Err().Op("!=").Nil()).Block(
				Qual(pkgUtils, "ExitOnError").Call(Err(), Lit("read configuration error")),
			),
		),
		Return(Id("configuration")),
	).Line()

	code.Line().Func().Id("Values").Params().Params(Id("config")).Block(
		Return(Op("*").Id("internalConfig").Call()),
	).Line()

	code.Line().Func().Id("getLinkedString").Params(Id("linked").Op("*").String()).Params(String()).Block(
		Line().If(Id("linked").Op("!=").Nil()).Block(
			Return(Op("*").Id("linked")),
		),
		Return(Lit("unset")),
	)

	code.Line().Func().Id("setLinkedString").Params(Id("linked").Op("**").String(), Id("value").String()).Block(
		Line().If(Op("*").Id("linked").Op("==").Nil()).Block(
			Op("*").Id("linked").Op("=").Op("&").Id("value"),
		),
	)
	return
}

func renderConfigStruct(meta metaInfo) Code {

	return Type().Id("config").StructFunc(func(g *Group) {

		for _, field := range fields {
			g.Id(field).Op("*").String()
		}

		g.Line().Comment("common env vars")
		g.Id("LogLevel").String().Tag(map[string]string{"env": "LOG_LEVEL", "envDefault": "debug"})
		g.Id("ServiceBind").String().Tag(map[string]string{"env": "BIND_ADDR", "envDefault": ":9000"})
		g.Id("PprofBind").String().Tag(map[string]string{"env": "BIND_PPROF", "envDefault": ":8080"})
		g.Id("HealthBind").String().Tag(map[string]string{"env": "BIND_HEALTH", "envDefault": ":9091"})
		g.Id("MetricsBind").String().Tag(map[string]string{"env": "BIND_METRICS", "envDefault": ":9090"})
		g.Id("EnablePPROF").Bool().Tag(map[string]string{"env": "ENABLE_PPROF", "envDefault": "false"})
		if meta.tracer == TracerZipkin {
			g.Id("Zipkin").String().Tag(map[string]string{"env": "ZIPKIN_ADDR", "envDefault": "https://zipkin.scnetservices.ru/api/v2/spans"})
		}

		g.Line().Comment(meta.projectName + " variables")

		if meta.withMongo {
			g.Id("MongoAddr").String().Tag(map[string]string{"env": "MONGO_ADDR", "envDefault": fmt.Sprintf("mongodb://mongo.default.svc.cluster.local/%s", meta.projectName)})
		}
	})
}
