package node

import (
	"github.com/seniorGolang/i2s/pkg/logger"
	"github.com/seniorGolang/i2s/pkg/tags"
)

type NodeParser struct {
	services []Service
	fileTags tags.DocTags

	types map[string]*Object
}

func New() *NodeParser {
	return &NodeParser{
		types: make(map[string]*Object),
	}
}

var log = logger.Log.WithField("module", "node")
