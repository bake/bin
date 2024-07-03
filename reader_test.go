package bin_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/bake/bin"
	"github.com/matryer/is"
)

func TestReadUint8(t *testing.T) {
	type Out struct {
		A uint8
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x00}, out: Out{A: 0}},
		{in: []byte{0x01}, out: Out{A: 1}},
		{in: []byte{0xff}, out: Out{A: 255}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadNested(t *testing.T) {
	type Value struct{ Value uint8 }

	type Out struct {
		A uint8
		B Value
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x00, 0x00}, out: Out{A: 0, B: Value{Value: 0}}},
		{in: []byte{0x01, 0x01}, out: Out{A: 1, B: Value{Value: 1}}},
		{in: []byte{0xff, 0x80}, out: Out{A: 255, B: Value{Value: 128}}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadFixedString(t *testing.T) {
	type Out struct {
		A string `bin:",size=4"`
		B string `bin:",size=3"`
		C string `bin:",size=5"`
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{
			0x72, 0x61, 0x74, 0x73,
			0x61, 0x72, 0x65,
			0x63, 0x75, 0x74, 0x65, 0x21,
		}, out: Out{A: "rats", B: "are", C: "cute!"}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadPrefixedString(t *testing.T) {
	type Number struct {
		Value uint8
	}

	type Out struct {
		Size  Number
		Value string `bin:",size=.Size.Value"`
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x04, 0x72, 0x61, 0x74, 0x73}, out: Out{Size: Number{Value: 4}, Value: "rats"}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadNestedPrefixedString(t *testing.T) {
	type Number struct {
		Value uint8
	}

	type Inner struct {
		Size  Number
		Value string `bin:",size=.Size.Value"`
	}

	type Out struct {
		Inner Inner
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x00}, out: Out{Inner{Size: Number{Value: 0}, Value: ""}}},
		{in: []byte{0x03, 0x61, 0x62, 0x63}, out: Out{Inner{Size: Number{Value: 3}, Value: "abc"}}},
		{in: []byte{0x04, 0x72, 0x61, 0x74, 0x73}, out: Out{Inner{Size: Number{Value: 4}, Value: "rats"}}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadArray(t *testing.T) {
	type Out struct {
		Numbers [3]uint8
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x00, 0x01, 0x02}, out: Out{Numbers: [3]uint8{0, 1, 2}}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadSlice(t *testing.T) {
	type Out struct {
		Numbers []uint8 `bin:",size=4"`
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x00, 0x01, 0x02, 0x3}, out: Out{Numbers: []uint8{0, 1, 2, 3}}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadPrefixedSlice(t *testing.T) {
	type Out struct {
		Size    uint8
		Numbers []uint8 `bin:",size=.Size"`
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x00}, out: Out{Size: 0, Numbers: []uint8{}}},
		{in: []byte{0x01, 0xff}, out: Out{Size: 1, Numbers: []uint8{255}}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadPrefixedSliceStruct(t *testing.T) {
	type Number struct {
		Value uint8
	}

	type Out struct {
		Size    uint8
		Numbers []Number `bin:",size=.Size"`
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x00}, out: Out{Size: 0, Numbers: []Number{}}},
		{in: []byte{0x01, 0xff}, out: Out{Size: 1, Numbers: []Number{{Value: 255}}}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}

func TestReadVarint(t *testing.T) {
	type Out struct {
		Number uint64 `bin:"uvarint"`
	}

	tt := []struct {
		in  []byte
		out Out
	}{
		{in: []byte{0x00}, out: Out{Number: 0}},
		{in: []byte{0x80, 0x08}, out: Out{Number: 1024}},
		{in: []byte{0xb9, 0x0a}, out: Out{Number: 1337}},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			is := is.New(t)
			r := bytes.NewReader(tc.in)
			br := bin.NewReader(r)
			var out Out
			err := br.Read(&out)
			is.NoErr(err)
			is.Equal(out, tc.out)
		})
	}
}
