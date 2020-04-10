package node

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/tags"
)

func (p *NodeParser) ParseEvents(eventsDirectory string) (events []Event, err error) {

	var files []os.FileInfo
	if files, err = ioutil.ReadDir(eventsDirectory); err != nil {
		return
	}

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		var goFile *types.File
		if goFile, err = astra.ParseFile(path.Join(eventsDirectory, file.Name())); err != nil {
			return
		}

		for _, object := range goFile.Structures {

			objectTags := tags.ParseTags(object.Docs)

			if objectTags.IsSet("event") {

				var objectType *Object
				objectType, err = p.parseObject(eventsDirectory, types.Variable{
					Base: object.Base,
					Type: object,
				})

				events = append(events, Event{
					Name: object.Name,
					Tags: objectTags,
					Type: objectType,
				})
			}
		}
	}
	return
}
