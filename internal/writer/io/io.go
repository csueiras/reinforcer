package io

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
)

// OutputProvider provides the means to write to the underlying storage medium such as a file system or an in-memory store
type OutputProvider interface {
	GetOutputTarget(filename string) (io.WriteCloser, error)
}

// FSOutputProvider is an OutputProvider that creates writers for the file system
type FSOutputProvider struct {
}

// NewFSOutputProvider creates an OutputProvider that creates writers for the file system
func NewFSOutputProvider() *FSOutputProvider {
	return &FSOutputProvider{}
}

// GetOutputTarget creates a local filesystem writer for the given filename, it will pre-create any directories that are
// in the filename's path
func (F *FSOutputProvider) GetOutputTarget(filename string) (io.WriteCloser, error) {
	dir := path.Dir(filename)
	if !path.IsAbs(dir) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dir = path.Join(cwd, path.Clean(dir))
	} else {
		dir = path.Clean(dir)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	filename = path.Base(filename)
	fullPath := path.Join(dir, filename)
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s; error=%w", fullPath, err)
	}
	return f, nil
}

type nopCloser struct {
	w io.Writer
}

func (f *nopCloser) Write(p []byte) (n int, err error) {
	return f.w.Write(p)
}

func (f *nopCloser) Close() error {
	return nil
}

// BufferOutputProvider is an OutputProvider that creates writers backed by bytes.Buffer
type BufferOutputProvider struct {
	Buffers map[string]*bytes.Buffer
}

// NewBufferOutputProvider is a constructor for the BufferOutputProvider
func NewBufferOutputProvider() *BufferOutputProvider {
	return &BufferOutputProvider{Buffers: map[string]*bytes.Buffer{}}
}

// GetOutputTarget creates a writer for an in memory bytes.Buffer uniquely identified by the given target argument
func (b *BufferOutputProvider) GetOutputTarget(target string) (io.WriteCloser, error) {
	b.Buffers[target] = &bytes.Buffer{}
	return &nopCloser{w: b.Buffers[target]}, nil
}
