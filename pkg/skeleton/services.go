package skeleton

import (
	"fmt"
	"os"
	"path"

	. "github.com/dave/jennifer/jen"
)

func genServices(meta metaInfo) (err error) {

	log.Info("generate services")

	if err = renderBaseService(meta, path.Join(meta.baseDir, "pkg", "service")); err != nil {
		return
	}

	typesPath := path.Join(meta.baseDir, "pkg", "service", "types")

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
	srcFile.PackageComment("@gkg version=0.0.1")
	srcFile.PackageComment(fmt.Sprintf("@gkg backend=%s", meta.projectName))
	srcFile.PackageComment(fmt.Sprintf("@gkg title=\"%s API\"", meta.projectName))
	srcFile.PackageComment(fmt.Sprintf("@gkg description=\"A service which provide %s API\"", meta.projectName))
	srcFile.PackageComment("@gkg servers=\"http://auth.k8s.platoon.dev.scnetservices.ru/api/v1;Sandbox for mobile developers|http://k8s.platoon.dev.scnetservices.ru/v1;description: Just for debug\"")

	srcFile.ImportName(pkgContext, "context")

	srcFile.Comment("@gkg jsonRPC-server log trace metrics test")
	srcFile.Type().Id("ServiceAPI").Interface(
		Id("Method").Params(Id("ctx").Qual(pkgContext, "Context")).Params(Err().Error()),
	)

	return srcFile.Save(path.Join(servicesPath, "service.go"))
}
