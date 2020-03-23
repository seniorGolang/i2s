package skeleton

import (
	"os"
	"os/exec"
	"path"

	"github.com/seniorGolang/i2s/pkg/server"
)

func GenerateSkeleton(projectName, repoName, baseDir string, jaeger, mongo bool) (err error) {

	meta := metaInfo{
		baseDir:     baseDir,
		repoName:    repoName,
		projectName: projectName,
		tracer:      TracerZipkin,
		withMongo:   mongo,
	}

	if jaeger {
		meta.tracer = TracerJaeger
	}

	log.Info("make directory")

	if err = os.MkdirAll(path.Join(baseDir, projectName), os.ModePerm); err != nil {
		return
	}

	if err = os.Chdir(path.Join(baseDir, projectName)); err != nil {
		return
	}

	log.Info("init go.mod")

	if err = exec.Command("go", "mod", "init", path.Join(meta.repoName, meta.projectName)).Run(); err != nil {
		return
	}

	if err = genConfig(meta); err != nil {
		return
	}

	if err = genServices(meta); err != nil {
		return
	}

	if err = server.MakeServices(path.Join(meta.baseDir, "pkg", "service"), "."); err != nil {
		return
	}

	if err = makeCmdMain(meta, path.Join(meta.repoName, meta.projectName), path.Join(meta.baseDir, "cmd", "service")); err != nil {
		return
	}

	log.Info("download dependencies ...")
	return exec.Command("go", "mod", "tidy").Run()
}
