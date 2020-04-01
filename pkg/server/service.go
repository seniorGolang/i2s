package server

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/logger"
	"github.com/seniorGolang/i2s/pkg/meta"
	"github.com/seniorGolang/i2s/pkg/tags"
	"github.com/seniorGolang/i2s/pkg/utils"
)

var log = logger.Log.WithField("module", "server")

func MakeServices(serviceDirectory, outPath string) (err error) {

	log.WithField("dir", serviceDirectory).Info("generate services transport")

	var files []os.FileInfo
	if files, err = ioutil.ReadDir(serviceDirectory); err != nil {
		return
	}

	var services []types.Interface

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		var serviceAst *types.File
		if serviceAst, err = astra.ParseFile(path.Join(serviceDirectory, file.Name())); err != nil {
			return
		}

		for _, iface := range serviceAst.Interfaces {
			if len(tags.ParseTags(iface.Docs)) != 0 {
				services = append(services, iface)
			}
		}
	}

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		if err = MakeService(path.Join(serviceDirectory, file.Name()), outPath, services); err != nil {
			return
		}
	}
	return
}

func MakeService(srcPath, outPath string, services []types.Interface) (err error) {

	var serviceAst *types.File

	if serviceAst, err = astra.ParseFile(srcPath); err != nil {
		return
	}

	for _, iface := range serviceAst.Interfaces {

		log.Infof("generate service %s transport", utils.ToLowerCamel(iface.Name))

		if err := validateInterface(iface); err != nil {
			log.Warning("iface", iface.Name, "problem", err.Error())
			continue
		}

		var info *meta.GenerationInfo

		if info, err = meta.FillInfo(&iface, srcPath, outPath); err != nil {
			return
		}

		info.Services = services

		toGenerate := make(map[string]func(*meta.GenerationInfo) error)

		ifaceTags := tags.ParseTags(iface.Docs)

		for key := range tags.ParseTags(iface.Docs) {

			switch key {

			case "log":
				toGenerate["renderServiceMiddlewareLogging"] = renderServiceMiddlewareLogging

				if !ifaceTags.IsSet("disableExchange") {
					toGenerate["renderTransportExchange"] = renderTransportExchange
				}

			case "test":
				toGenerate["renderServiceTests"] = renderServiceTests

			case "trace":
				toGenerate["renderServiceTracing"] = renderServiceTracing
				toGenerate["renderTransportEndpoints"] = renderTransportEndpoints

			case "metrics":
				toGenerate["renderServerMetrics"] = renderServerMetrics
				toGenerate["renderMetricsServe"] = renderMetricsServe
				toGenerate["renderServiceMiddlewareMetrics"] = renderServiceMiddlewareMetrics

			case "jsonRPC-server":
				toGenerate["renderHttpServer"] = renderHttpServer

				toGenerate["renderServerJsonRPC"] = renderServerJsonRPC
				toGenerate["renderServerTracing"] = renderServerTracing
				toGenerate["renderTransportJsonRPC"] = renderTransportJsonRPC
				toGenerate["renderTransportEndpoints"] = renderTransportEndpoints
				toGenerate["renderServerApplication"] = renderServerApplication

				if !ifaceTags.IsSet("disableEndpoints") {
					toGenerate["renderTransportServer"] = renderTransportServer
				}

				if !ifaceTags.IsSet("disableExchange") {
					toGenerate["renderTransportExchange"] = renderTransportExchange
				}

			case "http-server":
				toGenerate["renderHttpServer"] = renderHttpServer

				toGenerate["renderServerTracing"] = renderServerTracing
				toGenerate["renderServerApplication"] = renderServerApplication
				toGenerate["renderTransportEndpoints"] = renderTransportEndpoints
				toGenerate["renderTransportHttpServer"] = renderTransportHttpServer

				if !ifaceTags.IsSet("disableEndpoints") {
					toGenerate["renderTransportServer"] = renderTransportServer
				}

				if !ifaceTags.IsSet("disableExchange") {
					toGenerate["renderTransportExchange"] = renderTransportExchange
				}
			}
		}

		if len(toGenerate) > 0 {
			if err = os.MkdirAll(path.Join(outPath, "transport", strings.ToLower(info.ServiceName)), 0777); err != nil {
				return
			}
		}

		for _, gen := range toGenerate {

			if err = gen(info); err != nil {
				log.Error(err)
			}
		}
	}
	return
}
