package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/urfave/cli"

	"github.com/seniorGolang/i2s/pkg/logger"
	"github.com/seniorGolang/i2s/pkg/node"
	"github.com/seniorGolang/i2s/pkg/server"
	"github.com/seniorGolang/i2s/pkg/swagger"
)

var (
	GitSHA     = "-"
	BuildStamp = time.Now()
	Version    = "local.dev"
)

var log = logger.Log.WithField("module", "i2s")

func main() {

	app := cli.NewApp()
	app.Name = "Interface to Service Go-Kit Generator (i2s)"
	app.Usage = "make Go-Kit API easy"
	app.Version = Version
	if GitSHA != "-" {
		app.Version = Version + "-" + GitSHA
	}
	app.Compiled = BuildStamp
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:   "transport",
			Usage:  "generate services transport layer by interfaces in 'service' package",
			Action: cmdTransport,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "services",
					Value: "./pkg/someService/service",
					Usage: "path to services package",
				},
				cli.StringFlag{
					Name:  "out",
					Usage: "path to output folder",
				},
				cli.BoolFlag{
					Name:  "jaeger",
					Usage: "use Jaeger tracer",
				},
				cli.BoolFlag{
					Name:  "zipkin",
					Usage: "use Zipkin tracer",
				},
				cli.BoolFlag{
					Name:  "mongo",
					Usage: "enable mongo support",
				},
				cli.BoolFlag{
					Name:  "swagger",
					Usage: "generate swagger docs",
				},
			},

			UsageText:   "i2s transport",
			Description: "generate services transport layer by interfaces",
		},
		{
			Name:   "swagger",
			Usage:  "generate swagger documentation by interfaces in 'service' package",
			Action: cmdSwagger,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "services",
					Value: "./pkg/someService/service",
					Usage: "path to services package",
				},
				cli.StringFlag{
					Name:  "out",
					Usage: "path to output folder",
				},
				cli.StringSliceFlag{
					Name:  "iface",
					Usage: "interfaces included to swagger",
				},
				cli.BoolFlag{
					Name:  "json",
					Usage: "save swagger in JSON format",
				},
			},

			UsageText:   "i2s swagger --iface firstIface --iface secondIface",
			Description: "generate swagger documentation by interfaces",
		},
		{
			Name:   "meta",
			Usage:  "generate meta data in JSON",
			Action: cmdMeta,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "services",
					Value: "./pkg/someService/service",
					Usage: "path to services package",
				},
				cli.StringFlag{
					Name:  "out",
					Value: ".",
					Usage: "path to output folder",
				},
				cli.StringSliceFlag{
					Name:  "iface",
					Usage: "interfaces included to meta",
				},
			},

			UsageText:   "i2s meta --iface firstIface --iface secondIface",
			Description: "generate service meta data",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func cmdTransport(c *cli.Context) (err error) {

	defer func() {
		if err == nil {
			log.Info("done")
		}
	}()

	outPath := c.String("services")
	outPath, _ = path.Split(outPath + ".l")

	if c.String("out") != "" {
		outPath = c.String("out")
	}

	if err = server.MakeServices(c.String("services"), outPath); err != nil {
		return
	}

	if c.Bool("swagger") {
		err = buildSwagger(c.String("services"), outPath, false)
		if err != nil {
			return
		}
	}
	return
}

func cmdSwagger(c *cli.Context) (err error) {

	defer func() {
		if err == nil {
			log.Info("done")
		}
	}()

	outPath := c.String("services")
	outPath, _ = path.Split(outPath + ".ext")

	if c.String("out") != "" {
		outPath = c.String("out")
	}

	return buildSwagger(c.String("services"), outPath, c.Bool("json"), c.StringSlice("iface")...)
}

func cmdMeta(c *cli.Context) (err error) {

	defer func() {
		if err == nil {
			log.Info("done")
		}
	}()

	return buildMeta(
		c.String("services"),
		path.Join(c.String("out"), c.String("libName")),
		c.StringSlice("iface")...,
	)
}

func buildSwagger(ifaceFolder, outPath string, toJson bool, ifaceNames ...string) (err error) {

	nodeSource, err := node.Parse(ifaceFolder, ifaceNames...)

	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("parse %s for %v", ifaceFolder, ifaceNames))
	}

	var swaggerDoc swagger.Swagger
	if swaggerDoc, err = swagger.BuildSwagger(nodeSource); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("build swagger %s for %v", ifaceFolder, ifaceNames))
	}

	var swaggerData []byte
	fileName := path.Join(outPath, "swagger")

	if toJson {
		fileName += ".json"
		if swaggerData, err = json.Marshal(swaggerDoc); err != nil {
			return
		}
	} else {
		fileName += ".yaml"
		if swaggerData, err = yaml.Marshal(swaggerDoc); err != nil {
			return
		}
	}

	return ioutil.WriteFile(fileName, swaggerData, 0666)
}

func buildMeta(ifaceFolder string, outPath string, ifaceNames ...string) (err error) {

	var serviceInfo node.Node

	if serviceInfo, err = node.Parse(ifaceFolder, ifaceNames...); err != nil {
		return
	}

	return serviceInfo.SaveJSON(path.Join(outPath, "meta.json"))
}
