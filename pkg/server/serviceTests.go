package server

import (
	"context"
	"fmt"
	"path"
	"strings"

	. "github.com/dave/jennifer/jen"

	"github.com/seniorGolang/i2s/pkg/meta"
)

func renderServiceTests(info *meta.GenerationInfo) (err error) {

	log.Warn(info.OutputFilePath)

	srcFile := NewFileProxy(info.PkgName)

	ctx := prepareContext(info)
	ctx = context.WithValue(ctx, "code", srcFile)

	srcFile.ImportAlias(packageTesting, "testing")

	for _, method := range info.Iface.Methods {

		srcFile.Line().Func().Id(fmt.Sprintf("Test%s%s", info.Iface.Name, method.Name)).Params(Id("t").Op("*").Qual(packageTesting, "T")).Block(
			Id("t").Dot("Error").Call(Lit(fmt.Sprintf("test %s.%s is not implemented", info.Iface.Name, method.Name))),
		)

	}
	return srcFile.Save(path.Join(info.OutputFilePath, fmt.Sprintf("%s_test.go", strings.ToLower(info.ServiceName))))
}
