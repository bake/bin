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

func (t *Tag) Name() Kind {
	if t == nil || t.tag == nil {
		return ""
	}
	return Kind(t.tag.Name)
}

func (t *Tag) Option(key string) (string, bool) {
	if t == nil || t.tag == nil {
		return "", false
	}
	key += "="
	for _, o := range t.tag.Options {
		if !strings.HasPrefix(o, key) {
			continue
		}
		return strings.TrimPrefix(o, key), true
	}
	return "", false
}

func (t *Tag) HasOption(opt string) bool {
	if t == nil || t.tag == nil {
		return false
	}
	return t.tag.HasOption(opt)
}
