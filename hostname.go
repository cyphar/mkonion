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
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
)

const HostnamePath = "/var/lib/tor/hidden_service/hostname"

func GetOnionHostname(cli *client.Client, containerID string) (string, error) {
	content, stat, err := cli.CopyFromContainer(containerID, HostnamePath)
	// XXX: This isn't very pretty. But we need to wait until Tor generates
	//      an .onion address, and there's not really any better way of
	//      doing it.
	for err != nil && strings.Contains(err.Error(), "no such file or directory") {
		log.Warnf("tor onion hostname not found in container, retrying after a short nap...")
		time.Sleep(500 * time.Millisecond)
		content, stat, err = cli.CopyFromContainer(containerID, HostnamePath)
	}
	if err != nil {
		return "", err
	}
	defer content.Close()

	if stat.Mode.IsDir() {
		return "", fmt.Errorf("hostname file is a directory")
	}

	tr := tar.NewReader(content)
	hdr, err := tr.Next()
	for err != io.EOF {
		if err != nil {
			break
		}

		// XXX: Maybe do filepath.Base()?
		if hdr.Name != "hostname" {
			continue
		}

		data, err := ioutil.ReadAll(tr)
		if err != nil {
			return "", err
		}

		hostname := string(data)
		return strings.TrimSpace(hostname), nil
	}

	return "", fmt.Errorf("hostname file not in copied archive")
}
