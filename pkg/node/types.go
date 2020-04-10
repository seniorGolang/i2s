package node

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/seniorGolang/i2s/pkg/tags"
)

type Node struct {
	Name     string       `json:"name"`
	Tags     tags.DocTags `json:"tags,omitempty"`
	Events   []Event      `json:"events,omitempty"`
	Services []Service    `json:"services,omitempty"`
}

type Event struct {
	Name string       `json:"name"`
	Tags tags.DocTags `json:"tags,omitempty"`
	Type *Object      `json:"type,omitempty"`
}

type Service struct {
	Name    string       `json:"name"`
	Tags    tags.DocTags `json:"tags,omitempty"`
	Methods []Method     `json:"methods,omitempty"`
}

type Method struct {
	Name      string       `json:"name"`
	Tags      tags.DocTags `json:"tags,omitempty"`
	Results   []*Object    `json:"results,omitempty"`
	Arguments []*Object    `json:"arguments,omitempty"`

	HasContext  bool `json:"hasContext,omitempty"`
	ReturnError bool `json:"returnError,omitempty"`
}

type Object struct {
	Name   string    `json:"name"`
	Type   string    `json:"type"`
	Fields []*Object `json:"fields,omitempty"`

	Tags     tags.DocTags        `json:"tags,omitempty"`
	TypeTags map[string][]string `json:"typeTags,omitempty"`
	SubTypes map[string]*Object  `json:"subTypes,omitempty"`

	Alias string `json:"alias,omitempty"`

	IsMap      bool `json:"isMap,omitempty"`
	IsArray    bool `json:"isArray,omitempty"`
	IsBuildIn  bool `json:"isBuildIn,omitempty"`
	IsPrivate  bool `json:"isPrivate,omitempty"`
	IsNullable bool `json:"isNullable,omitempty"`
	IsEllipsis bool `json:"isEllipsis,omitempty"`
}

func (n Node) SaveJSON(path string) (err error) {

	var data []byte
	if data, err = json.MarshalIndent(n, "", " "); err != nil {
		return
	}

	return ioutil.WriteFile(path, data, 0666)
}

func (o Object) Value(str string) (value interface{}) {

	if o.Type == "bool" {

		value, _ = strconv.ParseBool(str)
		return

	} else if strings.HasPrefix(o.Type, "int") {

		value, _ = strconv.ParseInt(str, 10, 64)
		return

	} else if strings.HasPrefix(o.Type, "float") {

		value, _ = strconv.ParseFloat(str, 64)
		return
	}

	if str == "" || str == "null" {
		return nil
	}
	return str
}
