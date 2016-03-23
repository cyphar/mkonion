// mkonion: create a Tor onion service for existing Docker containers
// Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
	"github.com/docker/engine-api/types"
)

const HostnamePath = "/var/lib/tor/hidden_service/hostname"

func isRunning(state *types.ContainerState) bool {
	return state.Running && !state.Dead
}

func GetOnionHostname(cli *client.Client, containerID string) (string, error) {
	content, stat, err := cli.CopyFromContainer(containerID, HostnamePath)
	// XXX: This isn't very pretty. But we need to wait until Tor generates
	//      an .onion address, and there's not really any better way of
	//      doing it.
	for err != nil && strings.Contains(err.Error(), "no such file or directory") {
		// Make sure the container hasn't died.
		if inspect, err := cli.ContainerInspect(containerID); err != nil {
			return "", fmt.Errorf("error inspecting container: %s", err)
		} else if !isRunning(inspect.State) {
			return "", fmt.Errorf("container died before the hostname was computed")
		}

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
