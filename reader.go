package bin

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Kind string

const (
	Varint  Kind = "varint"
	UVarint Kind = "uvarint"
)

type Skipper interface {
	Skip(value reflect.Value) bool
}

type Sizer interface {
	Size(value reflect.Value) int
}

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
	return r.read(reflect.ValueOf(data), nil, "")
}

func (r *Reader) read(v reflect.Value, t *Tag, prefix string) error {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if cond, ok := t.Option("if"); ok {
		keep, err := r.cond(cond, prefix)
		if err != nil {
			return err
		}
		if !keep {
			return nil
		}
	}

	// TODO: Implementation differs from if.
	skip, err := r.skip(v, t, prefix)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	switch t.Name() {
	case Varint:
		return r.readVarint(v)
	case UVarint:
		return r.readUVarint(v)
	}

	switch v.Kind() {
	case reflect.Struct:
		return r.readStruct(v, t, prefix)
	case reflect.Uint8:
		return r.readUint8(v, t)
	case reflect.Uint16:
		return r.readUint16(v, t)
	case reflect.Uint32:
		return r.readUint32(v, t)
	case reflect.Uint64:
		return r.readUint64(v, t)
	case reflect.Int8:
		return r.readInt8(v, t)
	case reflect.Int16:
		return r.readInt16(v, t)
	case reflect.Int32:
		return r.readInt32(v, t)
	case reflect.Int64:
		return r.readInt64(v, t)
	case reflect.Float32:
		return r.readFloat32(v, t)
	case reflect.Float64:
		return r.readFloat64(v, t)
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

func (r *Reader) readVarint(v reflect.Value) error {
	out, err := binary.ReadVarint(r.reader)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(out).Convert(v.Type()))
	return nil
}

func (r *Reader) readUVarint(v reflect.Value) error {
	out, err := binary.ReadUvarint(r.reader)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(out).Convert(v.Type()))
	return nil
}

func (r *Reader) readStruct(v reflect.Value, t *Tag, prefix string) error {
	if size, eof, err := r.size(t, prefix); !eof && err == nil {
		lr := NewReader(io.LimitReader(r.reader, int64(size)))
		return lr.read(v, nil, prefix)
	}

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
	v.Set(reflect.ValueOf(uint8(buf[0])).Convert(v.Type()))
	return nil
}

func (r *Reader) readUint16(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 2)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	val := binary.LittleEndian.Uint16(buf)
	v.Set(reflect.ValueOf(val).Convert(v.Type()))
	return nil
}

func (r *Reader) readUint32(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 4)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	val := binary.LittleEndian.Uint32(buf)
	v.Set(reflect.ValueOf(val).Convert(v.Type()))
	return nil
}

func (r *Reader) readUint64(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 8)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	val := binary.LittleEndian.Uint64(buf)
	v.Set(reflect.ValueOf(val).Convert(v.Type()))
	return nil
}

func (r *Reader) readInt8(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 1)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	var n int8
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &n); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(n).Convert(v.Type()))
	return nil
}

func (r *Reader) readInt16(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 4)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	var n int16
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &n); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(n).Convert(v.Type()))
	return nil
}

func (r *Reader) readInt32(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 4)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	var n int32
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &n); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(n).Convert(v.Type()))
	return nil
}

func (r *Reader) readInt64(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 8)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	var n int64
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &n); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(n).Convert(v.Type()))
	return nil
}

func (r *Reader) readFloat32(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 4)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	bits := binary.LittleEndian.Uint32(buf)
	val := math.Float32frombits(bits)
	v.Set(reflect.ValueOf(val).Convert(v.Type()))
	return nil
}

func (r *Reader) readFloat64(v reflect.Value, _ *Tag) error {
	buf := make([]byte, 8)
	if _, err := r.reader.Read(buf); err != nil {
		return err
	}
	bits := binary.LittleEndian.Uint64(buf)
	val := math.Float64frombits(bits)
	v.Set(reflect.ValueOf(val).Convert(v.Type()))
	return nil
}

func (r *Reader) readString(v reflect.Value, t *Tag, prefix string) error {
	size, _, err := r.size(t, prefix)
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
	v.Set(reflect.ValueOf(string(buf)).Convert(v.Type()))
	return nil
}

func (r *Reader) readArray(v reflect.Value, _ *Tag, prefix string) error {
	for i := 0; i < v.Len(); i++ {
		name := fmt.Sprintf("%s[%d]", prefix, i)
		if err := r.read(v.Index(i), nil, name); err != nil {
			return errors.Wrapf(err, "could not read array value of %s", name)
		}
	}
	return nil
}

func (r *Reader) readSlice(v reflect.Value, t *Tag, prefix string) error {
	size, eof, err := r.size(t, prefix)
	if err != nil {
		return err
	}
	slice := reflect.MakeSlice(v.Type(), 0, size)
	for i := 0; i < size || eof; i++ {
		val := reflect.New(v.Type().Elem()).Elem()
		name := fmt.Sprintf("%s[%d]", prefix, i)
		err := r.read(val, nil, name)
		if eof && err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrapf(err, "could not read slice value of %s", name)
		}
		slice = reflect.Append(slice, val)
	}
	v.Set(slice)
	return nil
}

func (r *Reader) skip(v reflect.Value, t *Tag, prefix string) (bool, error) {
	field, ok := t.Option("skip")
	if !ok {
		return false, nil
	}
	skipper := reflect.TypeOf((*Skipper)(nil)).Elem()
	if !v.Type().Implements(skipper) {
		return false, fmt.Errorf("%q does not implement the Skipper interface", v.Type())
	}
	field = prefix[:strings.LastIndex(prefix, ".")] + field
	w, ok := r.fields[field]
	if !ok {
		return false, fmt.Errorf("could not find field %s", field)
	}
	if v.Interface().(Skipper).Skip(w) {
		return true, nil
	}
	return false, nil
}

func (r *Reader) size(t *Tag, prefix string) (int, bool, error) {
	size, ok := t.Option("size")
	if !ok {
		return 0, false, fmt.Errorf("size option not found")
	}

	if size == "EOF" {
		return 0, true, nil
	}

	if size, err := strconv.Atoi(size); err == nil {
		return size, false, nil
	}

	// Add the current fields prefix.
	size = prefix[:strings.LastIndex(prefix, ".")] + size

	if field, ok := r.fields[size]; ok {
		switch k := field.Kind(); k {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int(field.Uint()), false, nil
		default:
			return 0, false, fmt.Errorf("unknown field kind %q for %q", k, size)
		}
	}

	return 0, false, fmt.Errorf("could not parse size: %q", size)
}

func (r *Reader) cond(cond, prefix string) (bool, error) {
	fields := strings.Fields(cond)
	if len(fields) != 3 {
		return false, fmt.Errorf("unexpected if syntax: %q", cond)
	}

	left, ok := r.parseCond(fields[0], prefix)
	if !ok {
		return false, fmt.Errorf("could not parse left operant %q", fields[0])
	}

	right, ok := r.parseCond(fields[2], prefix)
	if !ok {
		return false, fmt.Errorf("could not parse right operant %q", fields[2])
	}

	switch fields[1] {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	}

	return false, nil
}

func (r *Reader) parseCond(cond, prefix string) (string, bool) {
	if cond[0] == '\'' && cond[len(cond)-1] == '\'' {
		return cond[1 : len(cond)-1], true
	}

	name := prefix[:strings.LastIndex(prefix, ".")] + cond
	if field, ok := r.fields[name]; ok {
		return field.String(), true
	}

	return "", false
}
