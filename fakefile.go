// mkonion: create a Tor onion service for existing Docker containers
// Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// 1. The above copyright notice and this permission notice shall be included in
//    all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"time"
)

type FakeFile struct {
	path string
	data []byte
	mode os.FileMode
}

func (ff *FakeFile) Name() string {
	return ff.path
}

func (ff *FakeFile) Size() int64 {
	return int64(len(ff.data))
}

func (ff *FakeFile) Mode() os.FileMode {
	return ff.mode
}

func (ff *FakeFile) ModTime() time.Time {
	// XXX: Does this break cache?
	return time.Now()
}

func (ff *FakeFile) IsDir() bool {
	return ff.Mode().IsDir()
}

func (ff *FakeFile) Sys() interface{} {
	return nil
}

func ArchiveContext(files []*FakeFile) (io.Reader, error) {
	archive := new(bytes.Buffer)
	tw := tar.NewWriter(archive)
	defer tw.Close()

	// XXX: What about directories?
	for _, file := range files {
		hdr, err := tar.FileInfoHeader(file, "")
		if err != nil {
			return nil, err
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}

		if _, err := tw.Write(file.data); err != nil {
			return nil, err
		}
	}

	tw.Flush()
	return bytes.NewReader(archive.Bytes()), nil
}
