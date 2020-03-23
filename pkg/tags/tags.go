package tags

import (
	"encoding/json"
	"strings"

	"github.com/seniorGolang/i2s/pkg/utils"
)

const (
	tagMark = "@i2s"
)

type DocTags map[string]string

func (tags DocTags) MarshalJSON() (bytes []byte, err error) {

	if len(tags) == 0 {
		return json.Marshal(nil)
	}
	return json.Marshal(map[string]string(tags))
}

func (tags DocTags) Merge(t DocTags) DocTags {

	if tags == nil {
		tags = make(DocTags)
	}

	for k, v := range t {
		tags[k] = v
	}
	return tags
}

func ParseTags(docs []string) (tags DocTags) {

	tags = make(DocTags)

	textLines := make(map[string][]string)

	for _, doc := range docs {

		doc = strings.TrimSpace(strings.TrimPrefix(doc, "//"))

		if strings.HasPrefix(doc, tagMark) {

			values, _ := TagScanner(doc[len(tagMark):])

			for k, v := range values {

				if _, found := tags[k]; found {
					tags[k] += "," + v
				} else {
					tags[k] = v
				}
			}
		}
	}

	for key, value := range tags {

		if !strings.HasPrefix(value, "#") {
			continue
		}

		for textKey, text := range textLines {
			if value == textKey {
				tags[key] = strings.Join(text, "\n")
			}
		}
	}
	return
}

func (tags DocTags) IsSet(tagName string) (found bool) {
	_, found = tags[tagName]
	return
}

func (tags DocTags) Contains(word string) (found bool) {

	for key := range tags {
		if strings.Contains(key, word) {
			return true
		}
	}
	return
}

func (tags DocTags) Value(tagName string, defValue ...string) (value string) {

	var found bool
	if value, found = tags[tagName]; !found {
		value = strings.Join(defValue, " ")
	}
	return
}

func (tags DocTags) ToKeys(tagName, separator string, defValue ...string) map[string]int {
	return utils.SliceStringToMap(strings.Split(tags.Value(tagName, defValue...), separator))
}

func (tags DocTags) ToMap(tagName, separator, splitter string, defValue ...string) (m map[string]string) {

	m = make(map[string]string)

	pairs := strings.Split(tags.Value(tagName, defValue...), separator)

	for _, pair := range pairs {
		if kv := strings.Split(pair, splitter); len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	return
}

func (tags DocTags) contains(tagName string) (found bool) {
	_, found = tags[tagName]
	return
}
