# bin

A Go reader for binary files. **This is work in progress.** It is not "complete" and may never be.

---

## Features

- Parse arrays and slices with arbitrary length. Read until EOF, a number that is known at build time or based on a previous struct field.
- Read variable length numbers (`varint` and `uvarint`).

## TODO

- Allow to cast `varint`and `uvarint` toany number type.
- Compare more than just `string`s.
- It is slow. Is shouldn't be.
- Document available struct tags.
- Allow for custom unmarshallers.

## Example

```go
func main() {
	type Character struct {
		NameLength uint8
		Name       string `bin:",size=.NameLength"`
	}

	type Out struct {
		Magic      string      `bin:",size=4"`
		Characters []Character `bin:",size=EOF"`
	}

	r := bytes.NewReader([]byte{
		0x4e, 0x69, 0x74, 0x57,
		0x03, 0x4d, 0x61, 0x65,
		0x03, 0x42, 0x65, 0x61,
		0x05, 0x47, 0x72, 0x65, 0x67, 0x67,
		0x05, 0x41, 0x6e, 0x67, 0x75, 0x73,
	})
	br := bin.NewReader(r)

	var out Out
	if err := br.Read(&out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, r := range out.Characters {
		fmt.Println(r.Name)
	}
}
```

See reader_test.go for more examples.
