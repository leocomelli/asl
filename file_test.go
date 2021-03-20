package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type Foo struct {
	ID   int
	Name string
}

const (
	tmpDir      = "/tmp"
	tmpFilename = "go-test"
)

func createTempFile(data []byte) ([]string, func()) {
	f, _ := os.CreateTemp(tmpDir, tmpFilename)
	f.Write(data)
	return filepath.SplitList(f.Name()), func() {
		os.Remove(f.Name())
	}
}

func TestReadFileAsByteArray(t *testing.T) {
	content := `
line1
line2
line3
`
	path, r := createTempFile([]byte(content))
	defer r()

	f := NewFile(path...)
	res, _ := f.Read()
	require.Equal(t, content, string(res))
}

func TestReadFileAsString(t *testing.T) {
	content := `
line1
line2
line3
`
	path, r := createTempFile([]byte(content))
	defer r()

	f := NewFile(path...)
	res, _ := f.ReadString()
	require.Equal(t, content, res)
}

func TestReadFileAsJSON(t *testing.T) {
	content := `
{
  "id": 1,
  "name": "foo"
}
`
	path, r := createTempFile([]byte(content))
	defer r()

	s := &Foo{}

	f := NewFile(path...)
	_ = f.ReadJSON(s)
	require.Equal(t, &Foo{ID: 1, Name: "foo"}, s)
}

func TestWriteTextFile(t *testing.T) {
	filename := "/tmp/go-test-text"
	defer os.Remove(filename)

	f := NewFile(filename)
	err := f.Write("foo")
	require.Nil(t, err)

	b, err := os.ReadFile(filename)
	require.Equal(t, "foo", string(b))
}

func TestWriteJSONFile(t *testing.T) {
	filename := "/tmp/go-test-text"
	defer os.Remove(filename)

	f := NewFile(filename)
	err := f.WriteJSON(&Foo{ID: 1, Name: "foo"})
	require.Nil(t, err)

	b, err := os.ReadFile(filename)
	require.Equal(t, "{\n \"ID\": 1,\n \"Name\": \"foo\"\n}", string(b))

}

func TestWriteTemplateInFile(t *testing.T) {
	filename := "/tmp/go-test-tmpl"
	defer os.Remove(filename)

	foo := &Foo{ID: 1, Name: "foo"}
	f := NewFile(filename)
	f.WriteTemplate("id: {{ .ID }}\nname: {{ .Name }}", foo)

	b, _ := os.ReadFile(filename)
	require.Equal(t, "id: 1\nname: foo", string(b))
}

func TestWriteTemplateSliceInFile(t *testing.T) {
	filename := "/tmp/go-test-tmplslice"
	defer os.Remove(filename)

	foo := []interface{}{&Foo{ID: 1, Name: "foo"}, &Foo{ID: 2, Name: "bar"}}
	f := NewFile(filename)
	f.WriteTemplateSlice("id: {{ .ID }}\nname: {{ .Name }}", foo)

	b, _ := os.ReadFile(filename)
	require.Equal(t, "id: 1\nname: fooid: 2\nname: bar", string(b))
}

func TestBackupFile(t *testing.T) {
	filename := "/tmp/go-test-bkp"
	os.WriteFile(filename, []byte("foo"), 0644)

	f := NewFile(filename)
	bkp, _ := f.Backup()

	require.True(t, strings.HasPrefix(bkp, filename))

	b, _ := os.ReadFile(filename)
	require.Equal(t, "foo", string(b))
}
