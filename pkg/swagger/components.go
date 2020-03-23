package swagger

import (
	"encoding/json"

	"github.com/seniorGolang/i2s/pkg/node"
	"github.com/seniorGolang/i2s/pkg/tags"
	"github.com/seniorGolang/i2s/pkg/utils"
)

const (
	formatHttp = iota
	formatJsonRPC
)

func addComponent(serviceName string, method node.Method, swagger *Swagger, format int) {

	if swagger.Components.Schemas == nil {
		swagger.Components.Schemas = make(map[string]schema)
	}

	requestType := "request" + utils.ToCamel(serviceName) + utils.ToCamel(method.Name)
	responseType := "response" + utils.ToCamel(serviceName) + utils.ToCamel(method.Name)

	var reqObject, respObject node.Object

	reqObject = node.Object{
		Name:   requestType,
		Type:   "object",
		Fields: method.Arguments,
	}

	respObject = node.Object{
		Name:   responseType,
		Type:   "object",
		Fields: method.Results,
	}

	if format == formatHttp {

		if method.Tags.IsSet("http-response-file") {
			respObject = node.Object{Name: "file", Type: "byte", IsArray: true}
		}
	}

	if format == formatJsonRPC {

		requestParams := node.Object{Name: "params", Fields: method.Arguments}

		reqObject = node.Object{
			Name: requestType,
			Type: "object",
			Fields: []node.Object{
				{Name: "id", Type: "uuid.UUID"},
				{Name: "jsonrpc", Type: "string", Tags: tags.DocTags{"example": "2.0"}},
			},
		}

		if len(method.Arguments) != 0 {
			reqObject.Fields = append(reqObject.Fields, requestParams)
		}

		responseResults := node.Object{Name: "result", Fields: method.Results}

		respObject = node.Object{
			Name: responseType,
			Type: "object",
			Fields: []node.Object{
				{Name: "id", Type: "uuid.UUID"},
				{Name: "jsonrpc", Type: "string", Tags: tags.DocTags{"example": "2.0"}},
			},
		}

		if len(method.Results) != 0 {
			respObject.Fields = append(respObject.Fields, responseResults)
		}
	}

	swagger.Components.Schemas[requestType] = makeComponent(reqObject)
	swagger.Components.Schemas[responseType] = makeComponent(respObject)
}

func makeComponent(object node.Object) (com schema) {

	com.Nullable = object.IsNullable
	com.Example = object.Value(object.Tags.Value("example"))
	com.Description = object.Tags.Value("desc")

	// if fakeType, found := object.TypeTags["fake"]; found && com.Example == "" {
	// 	fake.SetFake(fakeType[0], &com.Example)
	// }

	typeName, format := castType(object)

	if object.Type == "Interface" {

		com.Type = "object"
		if err := json.Unmarshal([]byte(object.Tags.Value("example", "{}")), &com.Example); err != nil {
			log.Error(err)
		}
		return
	}

	if object.IsArray {

		com.Type = "array"
		com.Items = &schema{Type: "object", Format: format, Properties: com.Properties}
		com.Properties = nil
		return
	}

	if object.IsMap {

		com.Type = "object"

		valueType, _ := castType(object.SubTypes["value"])

		com.AdditionalProperties = map[string]string{
			"type": valueType,
		}

		if err := json.Unmarshal([]byte(object.Tags.Value("example", "{}")), &com.Example); err != nil {
			log.Error(err)
		}
		return
	}

	if len(object.Fields) == 0 {

		if object.IsArray {
			com.Items = &schema{Type: typeName, Format: format}
			com.Type = "array"
			return
		}

		com.Type = typeName
		com.Format = format
		return
	}

	com.Type = "object"

	com.Properties = make(map[string]schema)

	for _, field := range object.Fields {

		if jsonTags, found := field.TypeTags["json"]; found {
			field.Name = jsonTags[0]
		}

		if !field.IsPrivate && field.Name != "-" {
			com.Properties[field.Name] = makeComponent(field)
		}
	}
	return
}

func castType(object node.Object) (typeName, format string) {

	typeName = object.Type

	switch typeName {

	case "bool":
		typeName = "boolean"

	case "Interface":
		typeName = "object"

	case "time.Time":
		format = "date-time"
		typeName = "string"

	case "byte":
		format = "byte"
		typeName = "string"

		if object.IsArray {
			format = "binary"
		}

	case "uuid.UUID":
		format = "uuid"
		typeName = "string"

	case "float32", "float64":
		format = "float"
		typeName = "number"

	case "int", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		typeName = "number"
	}

	format = object.Tags.Value("format", format)
	typeName = object.Tags.Value("type", typeName)

	return
}
