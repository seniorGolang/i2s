package swagger

import (
	"encoding/json"
	"io/ioutil"

	"github.com/seniorGolang/i2s/pkg/logger"
)

var (
	log = logger.Log.WithField("module", "swagger")
)

type Swagger struct {
	OpenAPI    string          `json:"openapi" yaml:"openapi"`
	Info       info            `json:"info,omitempty" yaml:"info,omitempty"`
	Servers    []server        `json:"servers,omitempty" yaml:"servers,omitempty"`
	Tags       []tag           `json:"tags,omitempty" yaml:"tags,omitempty"`
	Schemes    []string        `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Paths      map[string]path `json:"paths" yaml:"paths"`
	Components components      `json:"components,omitempty" yaml:"components,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type License struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

type info struct {
	Title          string   `json:"title,omitempty" yaml:"title,omitempty"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string   `json:"version,omitempty" yaml:"version,omitempty"`
}

type ExternalDocs struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
}

type tag struct {
	Name         string       `json:"name,omitempty" yaml:"name,omitempty"`
	Description  string       `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

type Properties map[string]schema

type schema struct {
	Ref         string      `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type        string      `json:"type,omitempty" yaml:"type,omitempty"`
	Format      string      `json:"format,omitempty" yaml:"format,omitempty"`
	Properties  Properties  `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items       *schema     `json:"items,omitempty" yaml:"items,omitempty"`
	Enum        []string    `json:"enum,omitempty" yaml:"enum,omitempty"`
	Nullable    bool        `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Example     interface{} `json:"example,omitempty" yaml:"example,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`

	OneOf []schema `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`

	AdditionalProperties interface{} `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
}

type parameter struct {
	Ref         string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	In          string `json:"in,omitempty" yaml:"in,omitempty"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool   `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type media struct {
	Schema schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type content map[string]media

type response struct {
	Description string  `json:"description" yaml:"description"`
	Content     content `json:"content,omitempty" yaml:"content,omitempty"`
}

type responses map[string]response

type requestBody struct {
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Content     content `json:"content,omitempty" yaml:"content,omitempty"`
	Required    bool    `json:"required,omitempty" yaml:"required,omitempty"`
}

type operation struct {
	Tags        []string     `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string       `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string       `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Consumes    []string     `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces    []string     `json:"produces,omitempty" yaml:"produces,omitempty"`
	Parameters  []parameter  `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *requestBody `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   responses    `json:"responses,omitempty" yaml:"responses,omitempty"`
}

type path struct {
	Ref         string     `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string     `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post        *operation `json:"post,omitempty" yaml:"post,omitempty"`
	Patch       *operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	Put         *operation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete      *operation `json:"delete,omitempty" yaml:"delete,omitempty"`
}

type variable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

type server struct {
	URL         string              `json:"url,omitempty" yaml:"url,omitempty"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]variable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type schemas map[string]schema

type components struct {
	Schemas schemas `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}

func (s *Swagger) SaveJSON(path string) (err error) {

	var data []byte
	if data, err = json.MarshalIndent(s, "", " "); err != nil {
		return
	}

	return ioutil.WriteFile(path, data, 0666)
}
