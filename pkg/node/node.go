package node

import (
	"github.com/seniorGolang/i2s/pkg/logger"
	"github.com/seniorGolang/i2s/pkg/tags"
)

type NodeParser struct {
	services []Service
	fileTags tags.DocTags

	objects map[string]*Object
}

func New() *NodeParser {
	return &NodeParser{
		objects: make(map[string]*Object),
	}
}

var log = logger.Log.WithField("module", "node")
