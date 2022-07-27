package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

const (
	filePerm os.FileMode = 0600
)

// File represents the file options
type File struct {
	Filename  string
	Path      string
	FullName  string
	Extension string
}

// NewFile returns a new File
func NewFile(path ...string) *File {
	p := filepath.Join(path...)
	return &File{
		Filename:  filepath.Base(p),
		Path:      filepath.Dir(p),
		FullName:  p,
		Extension: filepath.Ext(p),
	}
}

// Exists returns if file exists
func (f *File) Exists() bool {
	_, err := os.Stat(f.FullName)
	return !os.IsNotExist(err)
}

// Backup makes a copy of the file
func (f *File) Backup() (string, error) {
	if !f.Exists() {
		return "", nil
	}

	bkpFilename := fmt.Sprintf("%s.backup_%d", f.FullName, time.Now().Unix())
	b, err := os.ReadFile(f.FullName)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(bkpFilename, b, filePerm); err != nil {
		return "", err
	}

	return bkpFilename, nil
}

// Read reads a file and returns the content as []byte
func (f *File) Read() ([]byte, error) {
	b, err := os.ReadFile(f.FullName)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// ReadString reads a file and returns the content as string
func (f *File) ReadString() (string, error) {
	b, err := os.ReadFile(f.FullName)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ReadJSON reads a file and fills in the provided struct
func (f *File) ReadJSON(s interface{}) error {
	b, err := f.Read()
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, s)
	if err != nil {
		return err
	}

	return nil
}

// Write writes the string content in the file
func (f *File) Write(content string) error {
	return os.WriteFile(f.FullName, []byte(content), filePerm)
}

// WriteJSON writes the json in the file
func (f *File) WriteJSON(data interface{}) error {
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(f.FullName, b, filePerm)
}

// WriteTemplate writes the struct using a template
func (f *File) WriteTemplate(tmpl string, data interface{}) error {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err = t.Execute(&b, data); err != nil {
		return err
	}

	return os.WriteFile(f.FullName, b.Bytes(), filePerm)
}

// WriteTemplateSlice writes a slice using a template
func (f *File) WriteTemplateSlice(tmpl string, data []interface{}) error {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	for _, d := range data {
		if err = t.Execute(&b, d); err != nil {
			return err
		}
	}

	return os.WriteFile(f.FullName, b.Bytes(), filePerm)
}
