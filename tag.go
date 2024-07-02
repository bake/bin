package bin

import (
	"reflect"
	"strings"

	"github.com/fatih/structtag"
)

const (
	structTagKey = "bin"
)

type Tag struct {
	tag *structtag.Tag
}

func parseTags(st reflect.StructTag) (*Tag, error) {
	ts, err := structtag.Parse(string(st))
	if err != nil {
		return nil, err
	}
	t, err := ts.Get(structTagKey)
	if err != nil && err.Error() != "tag does not exist" {
		return nil, err
	}
	return &Tag{tag: t}, nil
}

func (t *Tag) Option(key string) (string, bool) {
	key += "="
	for _, o := range t.tag.Options {
		if !strings.HasPrefix(o, key) {
			continue
		}
		return strings.TrimPrefix(o, key), true
	}
	return "", false
}
