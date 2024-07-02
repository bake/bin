package bin

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Reader struct {
	reader io.Reader
	fields map[string]reflect.Value
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		reader: r,
		fields: map[string]reflect.Value{},
	}
}

func (r *Reader) Read(data any) error {
	t, err := parseTags(`bin:","`)
	if err != nil {
		return err
	}
	if err := r.read(reflect.ValueOf(data), t, ""); err != nil {
		return err
	}
	return nil
}

func (r *Reader) read(v reflect.Value, t *Tag, prefix string) error {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return r.readStruct(v, t, prefix)
	case reflect.Uint8:
		return r.readUint8(v, t)
	case reflect.String:
		return r.readString(v, t, prefix)
	default:
		fmt.Println("unknown kind", v.Kind())
	}
	return nil
}

func (r *Reader) readStruct(v reflect.Value, _ *Tag, prefix string) error {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := fmt.Sprintf("%s.%s", prefix, v.Type().Field(i).Name)
		tag, err := parseTags(v.Type().Field(i).Tag)
		if err != nil {
			return err
		}
		r.fields[name] = field
		if err := r.read(field, tag, name); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) readUint8(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 1)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(uint8(buf[0])))
	return nil
}

func (r *Reader) readString(v reflect.Value, t *Tag, prefix string) error {
	size, err := r.size(t, prefix)
	if err != nil {
		return err
	}
	if size == 0 {
		return nil
	}
	buf := make([]byte, size)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(string(buf)))
	return nil
}

func (r *Reader) size(t *Tag, prefix string) (int, error) {
	o, ok := t.Option("size")
	if !ok {
		return 0, fmt.Errorf("strings require a size option")
	}

	if size, err := strconv.Atoi(o); err == nil {
		return size, nil
	}

	// Add the current fields prefix.
	o = prefix[:strings.LastIndex(prefix, ".")] + o

	if field, ok := r.fields[o]; ok {
		switch k := field.Kind(); k {
		case reflect.Uint8:
			return int(field.Uint()), nil
		default:
			return 0, fmt.Errorf("unknown field kind %q for %q", k, o)
		}
	}

	return 0, fmt.Errorf("could not parse size")
}
