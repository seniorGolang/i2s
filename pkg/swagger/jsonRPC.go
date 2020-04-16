package swagger

import (
	"fmt"
	"strings"

	"github.com/seniorGolang/i2s/pkg/node"
	"github.com/seniorGolang/i2s/pkg/tags"
	"github.com/seniorGolang/i2s/pkg/utils"
)

func (b *Builder) buildJsonRPC(node node.Node, swagger *Swagger) {

	for _, service := range node.Services {

		if service.Tags.IsSet("jsonRPC-server") {

			log.Infof("generate swagger by service %s (jsonRPC transport layer)", service.Name)

			b.addErrorJsonRPC(swagger)

			for _, method := range service.Methods {

				op := &operation{
					Description: method.Tags.Value("desc"),
					Summary:     method.Tags.Value("summary"),
					Tags:        []string{utils.ToLowerCamel(service.Name)},
				}

				b.moveArgumentsToParameters(&method, op, swagger)
				b.addRequestResponse(service.Name, method, swagger, formatJsonRPC)

				httpPath := fmt.Sprintf("/%s/%s", utils.ToLowerCamel(service.Name), utils.ToLowerCamel(method.Name))
				httpReqContentTypes := strings.Split(method.Tags.Value("http-request-content-type", "application/json"), "|")
				httpResContentTypes := strings.Split(method.Tags.Value("http-response-content-type", "application/json"), "|")

				requestContent := make(content)

				for _, c := range httpReqContentTypes {

					requestContent[c] = media{
						Schema: schema{
							Ref: fmt.Sprintf("#/components/schemas/request%s%s", utils.ToCamel(service.Name), utils.ToCamel(method.Name)),
						},
					}
				}

				responseContent := make(content)

				for _, c := range httpResContentTypes {

					responseContent[c] = media{
						Schema: schema{
							Ref: fmt.Sprintf("#/components/schemas/response%s%s", utils.ToCamel(service.Name), utils.ToCamel(method.Name)),
						},
					}
				}

				op.RequestBody = &requestBody{
					Description: "",
					Required:    false,
					Content: content{
						"application/json": media{
							Schema: schema{
								Ref: fmt.Sprintf("#/components/schemas/request%s%s", utils.ToCamel(service.Name), utils.ToCamel(method.Name)),
							},
						},
					},
				}

				op.Responses = responses{
					"200": response{
						Description: "success or error",
						Content: content{
							"application/json": media{
								Schema: schema{
									OneOf: []schema{
										{Ref: fmt.Sprintf("#/components/schemas/response%s%s", utils.ToCamel(service.Name), utils.ToCamel(method.Name))},
										{Ref: "#/components/schemas/errorJsonRPC"},
									},
								},
							},
						},
					},
				}

				swagger.Paths[httpPath] = path{Post: op}
			}
		}
	}
}

func (b *Builder) addErrorJsonRPC(swagger *Swagger) {

	if swagger.Components.Schemas == nil {
		swagger.Components.Schemas = make(map[string]schema)
	}

	swagger.Components.Schemas["errorJsonRPC"] = b.makeType(&node.Object{
		Alias: "-",
		Name:  "errorJsonRPC",
		Type:  "object",
		Fields: []*node.Object{
			{Name: "id", Type: "uuid.UUID"},
			{Name: "jsonrpc", Type: "string", Tags: tags.DocTags{"example": "2.0"}},
			{Name: "error", Alias: "-", Fields: []*node.Object{
				{Name: "code", Type: "int", Tags: tags.DocTags{"example": "-32603"}},
				{Name: "message", Type: "string", Tags: tags.DocTags{"example": "not found"}},
				{Name: "data", Type: "object", IsNullable: true},
			}},
		},
	}, swagger)
}
