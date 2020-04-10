package swagger

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/seniorGolang/i2s/pkg/node"
	"github.com/seniorGolang/i2s/pkg/utils"
)

type move struct {
	pType string
	name  string
}

func (b *Builder) buildHttp(node node.Node, swagger *Swagger) {

	for _, service := range node.Services {

		if service.Tags.IsSet("http-server") {

			log.Infof("generate swagger by service %s (HTTP transport layer)", service.Name)

			for _, method := range service.Methods {

				var httpReqContentTypes, httpResContentTypes []string

				httpMethod := method.Tags.Value("http-method", "POST")
				httpPath := "/" + method.Tags.Value("http-path", utils.ToLowerCamel(method.Name))

				op := &operation{
					Description: method.Tags.Value("desc"),
					Summary:     method.Tags.Value("summary"),
					Tags:        []string{utils.ToLowerCamel(service.Name)},
				}

				b.moveArgumentsToParameters(&method, op, swagger)
				b.addRequestResponse(service.Name, method, swagger, formatHttp)

				if len(method.Arguments) != 0 {
					httpReqContentTypes = strings.Split(method.Tags.Value("http-request-content-type", "application/json"), "|")
				}

				if len(method.Results) != 0 {
					httpResContentTypes = strings.Split(method.Tags.Value("http-response-content-type", "application/json"), "|")
				}

				var requestContent, responseContent content

				for _, c := range httpReqContentTypes {

					if requestContent == nil {
						requestContent = make(content)
					}

					requestContent[c] = media{
						Schema: schema{
							Ref: fmt.Sprintf("#/components/schemas/request%s%s", utils.ToCamel(service.Name), utils.ToCamel(method.Name)),
						},
					}
				}

				for _, c := range httpResContentTypes {

					if responseContent == nil {
						responseContent = make(content)
					}

					responseContent[c] = media{
						Schema: schema{
							Ref: fmt.Sprintf("#/components/schemas/response%s%s", utils.ToCamel(service.Name), utils.ToCamel(method.Name)),
						},
					}
				}

				if requestContent != nil {
					op.RequestBody = &requestBody{
						Content: requestContent,
					}
				}

				op.Responses = responses{
					"200": response{
						Content: responseContent,
					},
				}

				httpCodeDesc := strings.Split(method.Tags.Value("http-code-descriptions"), ",")

				for _, codeDesc := range httpCodeDesc {

					if values := strings.Split(codeDesc, "|"); len(values) == 2 {

						if res, found := op.Responses[values[0]]; found {
							res.Description = values[1]
							op.Responses[values[0]] = res
							continue
						}
						op.Responses[values[0]] = response{Description: values[1]}
					}
				}

				var httpValue path
				httpValue, _ = swagger.Paths[httpPath]

				reflect.ValueOf(&httpValue).Elem().FieldByName(utils.ToCamel(strings.ToLower(httpMethod))).Set(reflect.ValueOf(op))
				swagger.Paths[httpPath] = httpValue
			}
		}
	}
}

func (b *Builder) moveArgumentsToParameters(method *node.Method, op *operation, swagger *Swagger) {

	pMove := make(map[string]move)

	headerParams := strings.Split(method.Tags.Value("http-headers"), ",")

	for _, headerParam := range headerParams {
		if pair := strings.Split(headerParam, "|"); len(pair) == 2 {

			argName := pair[0]
			headerName := strings.ToLower(strings.TrimSpace(pair[1]))

			pMove[argName] = move{pType: "header", name: headerName}
		}
	}

	argParams := strings.Split(method.Tags.Value("http-arg"), ",")

	for _, argParam := range argParams {
		pMove[argParam] = move{pType: "query", name: argParam}
	}

	var urlVars []string
	urlTokens := strings.Split(method.Tags.Value("http-path"), "/")

	for _, token := range urlTokens {
		if strings.HasPrefix(token, "{") {
			urlVars = append(urlVars, strings.TrimSpace(strings.Replace(strings.TrimPrefix(token, "{"), "}", "", -1)))
		}
	}

	for _, argParam := range urlVars {
		pMove[argParam] = move{pType: "path", name: argParam}
	}

	var clearArguments []*node.Object
	for _, argObject := range method.Arguments {

		if move, found := pMove[argObject.Name]; found {

			param := parameter{
				Name:     move.name,
				In:       move.pType,
				Required: true,
				Schema:   b.makeType(argObject, swagger),
			}

			if move.pType == "query" {
				param.Required = false
			}

			op.Parameters = append(op.Parameters, param)

		} else {
			clearArguments = append(clearArguments, argObject)
		}
	}
	method.Arguments = clearArguments
}
