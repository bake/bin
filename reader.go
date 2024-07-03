package bin

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Reader struct {
	reader *bufio.Reader
	fields map[string]reflect.Value
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(r),
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

	switch t.tag.Name {
	case "varint":
		return r.readVarint(v, t, prefix)
	case "uvarint":
		return r.readUVarint(v, t, prefix)
	}

	switch v.Kind() {
	case reflect.Struct:
		return r.readStruct(v, t, prefix)
	case reflect.Uint8:
		return r.readUint8(v, t)
	case reflect.String:
		return r.readString(v, t, prefix)
	case reflect.Array:
		return r.readArray(v, t, prefix)
	case reflect.Slice:
		return r.readSlice(v, t, prefix)
	default:
		return fmt.Errorf("unknown kind %q", v.Kind())
	}
}

// TODO: Test varint
func (r *Reader) readVarint(v reflect.Value, t *Tag, prefix string) error {
	out, err := binary.ReadVarint(r.reader)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(out))
	return nil
}

func (r *Reader) readUVarint(v reflect.Value, t *Tag, prefix string) error {
	out, err := binary.ReadUvarint(r.reader)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(out))
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

func (r *Reader) readArray(v reflect.Value, _ *Tag, prefix string) error {
	for i := 0; i < v.Len(); i++ {
		// TODO: Not used
		name := fmt.Sprintf("%s[%d]", prefix, i)
		if err := r.read(v.Index(i), nil, name); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) readSlice(v reflect.Value, t *Tag, prefix string) error {
	size, err := r.size(t, prefix)
	if err != nil {
		return err
	}
	v.Set(reflect.MakeSlice(v.Type(), size, size))
	for i := 0; i < v.Len(); i++ {
		// TODO: Not used
		name := fmt.Sprintf("%s[%d]", prefix, i)
		if err := r.read(v.Index(i), nil, name); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) size(t *Tag, prefix string) (int, error) {
	size, ok := t.Option("size")
	if !ok {
		return 0, fmt.Errorf("size option not found")
	}

	if size, err := strconv.Atoi(size); err == nil {
		return size, nil
	}

	// Add the current fields prefix.
	size = prefix[:strings.LastIndex(prefix, ".")] + size

	if field, ok := r.fields[size]; ok {
		switch k := field.Kind(); k {
		case reflect.Uint8:
			return int(field.Uint()), nil
		default:
			return 0, fmt.Errorf("unknown field kind %q for %q", k, size)
		}
	}

	return 0, fmt.Errorf("could not parse size")
}
