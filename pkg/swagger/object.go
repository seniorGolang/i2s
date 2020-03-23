package swagger

import (
	"github.com/seniorGolang/i2s/pkg/node"
	"github.com/seniorGolang/i2s/pkg/utils"
)

func buildObjects(node node.Node, swagger *Swagger) {

	for _, service := range node.Services {

		if service.Tags.IsSet("jsonRPC-server") {

			for _, method := range service.Methods {

				op := &operation{
					Description: method.Tags.Value("desc"),
					Summary:     method.Tags.Value("summary"),
					Tags:        []string{utils.ToLowerCamel(service.Name)},
				}

				moveArgumentsToParameters(&method, op)
			}
		}
	}
}
