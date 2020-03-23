package swagger

import (
	"github.com/seniorGolang/i2s/pkg/node"
)

func BuildSwagger(node node.Node) (swagger Swagger, err error) {

	makeSwagger(node, &swagger)

	buildHttp(node, &swagger)
	buildJsonRPC(node, &swagger)

	return
}
