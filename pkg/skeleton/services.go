package skeleton

import (
	"fmt"
	"os"
	"path"

	. "github.com/dave/jennifer/jen"

	"github.com/seniorGolang/i2s/pkg/utils"
)

func genServices(meta metaInfo) (err error) {

	log.Info("generate services")

	if err = renderBaseService(meta, path.Join(meta.baseDir, "pkg", meta.projectName, "service")); err != nil {
		return
	}

	typesPath := path.Join(meta.baseDir, "pkg", meta.projectName, "service", "types")

	if err = os.MkdirAll(typesPath, os.ModePerm); err != nil {
		return
	}
	return
}

func renderBaseService(meta metaInfo, servicesPath string) (err error) {

	if err = os.MkdirAll(servicesPath, os.ModePerm); err != nil {
		return
	}

	srcFile := NewFile("service")
	srcFile.PackageComment("@i2s version=0.0.1")
	srcFile.PackageComment(fmt.Sprintf("@i2s backend=%s", meta.projectName))
	srcFile.PackageComment(fmt.Sprintf("@i2s title=\"%s API\"", meta.projectName))
	srcFile.PackageComment(fmt.Sprintf("@i2s description=`A service which provide %s API`", meta.projectName))
	srcFile.PackageComment(fmt.Sprintf("@i2s servers=`http://%s:9000`", meta.projectName))

	srcFile.ImportName(pkgContext, "context")

	srcFile.Comment("@i2s jsonRPC-server log trace metrics test")
	srcFile.Type().Id(utils.ToCamel(meta.projectName)).Interface(
		Id("Method").Params(Id("ctx").Qual(pkgContext, "Context")).Params(Err().Error()),
	)

	return srcFile.Save(path.Join(servicesPath, fmt.Sprintf("%s.go", meta.projectName)))
}
