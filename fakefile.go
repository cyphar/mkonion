// mkonion: create a Tor onion service for existing Docker containers
// Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
