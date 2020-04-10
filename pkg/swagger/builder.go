package swagger

import (
	"github.com/seniorGolang/i2s/pkg/node"
)

func (b *Builder) BuildSwagger(node node.Node) (swagger Swagger, err error) {

	b.makeSwagger(node, &swagger)

	b.buildTypes(node, &swagger)

	b.buildHttp(node, &swagger)
	b.buildJsonRPC(node, &swagger)
	return
}
