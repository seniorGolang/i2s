package swagger

import (
	"strings"

	"github.com/seniorGolang/i2s/pkg/node"
)

func (b *Builder) makeSwagger(node node.Node, swagger *Swagger) {

	log.Info("generate service info")

	swagger.OpenAPI = "3.0.0"
	swagger.Info.Title = node.Tags.Value("title")
	swagger.Info.Version = node.Tags.Value("version")
	swagger.Info.Description = node.Tags.Value("description")
	swagger.Paths = make(map[string]path)

	tagServers := strings.Split(node.Tags.Value("servers"), "|")

	// fill servers
	for _, tagServer := range tagServers {

		var serverDescr string
		serverValues := strings.Split(tagServer, ";")

		if len(serverValues) > 1 {
			serverDescr = serverValues[1]
		}
		swagger.Servers = append(swagger.Servers, server{URL: serverValues[0], Description: serverDescr})
	}
}
