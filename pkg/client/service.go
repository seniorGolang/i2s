package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/logger"
	"github.com/seniorGolang/i2s/pkg/meta"
	"github.com/seniorGolang/i2s/pkg/tags"
)

var log = logger.Log.WithField("module", "client")

func MakeServices(serviceDirectory, outPath string) (err error) {

	var files []os.FileInfo
	if files, err = ioutil.ReadDir(serviceDirectory); err != nil {
		return
	}

	var backend string

	var services []types.Interface
	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		var serviceAst *types.File
		if serviceAst, err = astra.ParseFile(path.Join(serviceDirectory, file.Name())); err != nil {
			return
		}

		if backendName := tags.ParseTags(serviceAst.Docs).Value("backend", ""); backendName != "" {
			backend = backendName
		}

		for _, iface := range serviceAst.Interfaces {
			services = append(services, iface)
		}
	}

	if backend == "" {
		err = fmt.Errorf("backend var is not set")
		return
	}

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		if err = MakeService(backend, path.Join(serviceDirectory, file.Name()), outPath, services); err != nil {
			return
		}
	}
	return
}

func MakeService(backend, srcPath, outPath string, services []types.Interface) (err error) {

	var serviceAst *types.File

	if serviceAst, err = astra.ParseFile(srcPath); err != nil {
		return
	}

	for _, iface := range serviceAst.Interfaces {

		if !tags.ParseTags(iface.Docs).IsSet("jsonRPC-server") {
			continue
		}

		if err := validateInterface(iface); err != nil {
			log.Warning("iface", iface.Name, "problem", err.Error())
			continue
		}

		var info *meta.GenerationInfo

		if info, err = meta.FillInfo(&iface, srcPath, outPath); err != nil {
			return
		}

		info.Backend = backend
		info.Services = services

		if err = os.MkdirAll(path.Join(outPath, strings.ToLower(info.ServiceName)), 0777); err != nil {
			return
		}

		err = renderTransportClient(info)
		if err != nil {
			return
		}
		err = renderTransportEndpoints(info)
		if err != nil {
			return
		}
		err = renderTransportExchange(info)
		if err != nil {
			return
		}
		err = renderClientJsonRPC(info)
		if err != nil {
			return
		}
	}
	return
}
