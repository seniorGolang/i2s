package skeleton

import (
	"os/exec"
	"path"

	"github.com/seniorGolang/i2s/pkg/server"
)

func GenerateSkeleton(projectName, repoName, baseDir string, jaeger, zipkin, mongo bool) (err error) {

	meta := metaInfo{
		baseDir:     baseDir,
		repoName:    repoName,
		projectName: projectName,
		withMongo:   mongo,
	}

	if jaeger {
		meta.tracer = TracerJaeger
	}
	if zipkin {
		meta.tracer = TracerZipkin
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

	if err = server.MakeServices(path.Join(meta.baseDir, "pkg", projectName, "service"), path.Join(meta.baseDir, "pkg", projectName)); err != nil {
		return
	}

	if err = makeCmdMain(meta, path.Join(meta.repoName, meta.projectName), path.Join(meta.baseDir, "cmd", projectName)); err != nil {
		return
	}

	log.Info("download dependencies ...")
	return exec.Command("go", "mod", "tidy").Run()
}
