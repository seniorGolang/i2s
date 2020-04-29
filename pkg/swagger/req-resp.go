package swagger

import (
	"encoding/json"
	"fmt"

	"github.com/seniorGolang/i2s/pkg/node"
	"github.com/seniorGolang/i2s/pkg/tags"
	"github.com/seniorGolang/i2s/pkg/utils"
)

const (
	formatHttp = iota
	formatJsonRPC
)

func (b *Builder) addRequestResponse(serviceName string, method node.Method, swagger *Swagger, format int) {

	if swagger.Components.Schemas == nil {
		swagger.Components.Schemas = make(map[string]schema)
	}

	requestType := "request" + utils.ToCamel(serviceName) + utils.ToCamel(method.Name)
	responseType := "response" + utils.ToCamel(serviceName) + utils.ToCamel(method.Name)

	var reqObject, respObject *node.Object

	reqObject = &node.Object{
		Name: requestType,
		Type: "object",
	}

	respObject = &node.Object{
		Name: responseType,
		Type: "object",
	}

	if format == formatHttp {

		if method.Tags.IsSet("http-response-file") {
			respObject = &node.Object{Name: "file", Type: "byte", IsArray: true}
		}

		if len(method.Results) > 0 {
			swagger.Components.Schemas[responseType] = b.makeComponent(method.Results, swagger)
		}

		if len(method.Arguments) > 0 {
			swagger.Components.Schemas[requestType] = b.makeComponent(method.Arguments, swagger)
		}
	}

	if format == formatJsonRPC {

		reqObject = &node.Object{
			Alias: "-",
			Name:  requestType,
			Type:  "object",
			Fields: []*node.Object{
				{Name: "id", Type: "uuid.UUID"},
				{Name: "jsonrpc", Type: "string", Tags: tags.DocTags{"example": "2.0"}},
			},
		}

		respObject = &node.Object{
			Alias: "-",
			Name:  responseType,
			Type:  "object",
			Fields: []*node.Object{
				{Name: "id", Type: "uuid.UUID"},
				{Name: "jsonrpc", Type: "string", Tags: tags.DocTags{"example": "2.0"}},
			},
		}

		swagger.Components.Schemas[requestType] = b.makeType(reqObject, swagger)
		swagger.Components.Schemas[responseType] = b.makeType(respObject, swagger)

		if len(method.Results) > 0 {
			swagger.Components.Schemas[responseType].Properties["result"] = b.makeComponent(method.Results, swagger)
		}

		if len(method.Arguments) > 0 {
			if swagger.Components.Schemas[requestType].Properties == nil {
				s := swagger.Components.Schemas[requestType]
				s.Properties = Properties{"params": b.makeComponent(method.Arguments, swagger)}
				swagger.Components.Schemas[requestType] = s
			}
			swagger.Components.Schemas[requestType].Properties["params"] = b.makeComponent(method.Arguments, swagger)
		}
	}
}

func (b *Builder) makeComponent(fields []*node.Object, swagger *Swagger) (com schema) {

	com.Properties = make(Properties)

	for _, field := range fields {

		com.Nullable = field.IsNullable
		com.Example = field.Value(field.Tags.Value("example"))
		com.Description = field.Tags.Value("desc")

		if field.Type == "Interface" {

			com.Properties[field.Name] = schema{Type: "object"}
			if err := json.Unmarshal([]byte(field.Tags.Value("example", "{}")), &com.Example); err != nil {
				log.Error(err)
			}
			continue
		}

		if field.Alias != "" && field.Alias != "-" {

			if field.IsArray {
				com.Properties[field.Name] = schema{Type: "array", Items: &schema{Ref: fmt.Sprintf("#/components/schemas/%s", field.Alias)}}
				continue
			}

			com.Properties[field.Name] = schema{Ref: fmt.Sprintf("#/components/schemas/%s", field.Alias)}
			continue
		}

		if len(field.Fields) != 0 {
			com.Properties[field.Name] = schema{Ref: fmt.Sprintf("#/components/schemas/%s", field.Type)}
		}

		typeName, format := castType(field)

		if len(field.Fields) == 0 {

			if field.IsArray {
				com.Properties[field.Name] = schema{Type: "array", Format: format}
				continue
			}

			if field.IsMap {

				typeNameVal, formatVal := castType(field.SubTypes["value"])
				additionalProperties := schema{Type: typeNameVal, Format: formatVal}

				if len(field.SubTypes["value"].Fields) > 0 {
					additionalProperties = schema{Ref: fmt.Sprintf("#/components/schemas/%s", typeNameVal)}
					swagger.Components.Schemas[typeNameVal] = b.makeType(field.SubTypes["value"], swagger)
				}

				if field.SubTypes["value"].Alias != "" {
					additionalProperties = schema{Ref: fmt.Sprintf("#/components/schemas/%s", field.SubTypes["value"].Alias)}
				}

				com.Properties[field.Name] = schema{Type: "object", AdditionalProperties: additionalProperties}
				continue
			}

			com.Properties[field.Name] = schema{Type: typeName, Format: format}
		}
	}
	return
}
