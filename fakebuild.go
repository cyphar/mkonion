// mkonion: create a Tor onion service for existing Docker containers
// Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	containerTypes "github.com/docker/engine-api/types/container"
)

// Building a Docker image usually requires a real filesystem in order to create
// the tar archive sent to the Docker daemon. But if all you have a single
// configuration file, it's a pain to do that. In a sense, it's much nicer to
// embed all of the required information into the `mkonion` binary so we don't
// have to touch the filesystem.

const (
	MkonionTag                = "mkonion/tor:latest"
	MkonionDockerfileTemplate = `
	FROM alpine:latest
	RUN echo '@testing http://nl.alpinelinux.org/alpine/edge/testing' >> /etc/apk/repositories && \
		apk update && \
		apk upgrade && \
	    apk add --update \
			tor@testing && \
		{{ if .HasKey }}
		mkdir -p /var/lib/tor/hidden_service && \
		chmod 600 /var/lib/tor/hidden_service && \
		{{ end }}
		rm -rf /var/cache/apk/*
	COPY torrc /etc/tor/torrc
	{{ if .HasKey }}
	COPY private_key /var/lib/tor/hidden_service/private_key
	{{ end }}
	ENTRYPOINT ["/usr/bin/tor", "-f", "/etc/tor/torrc"]
	`
)

var dockerfileTemplate = template.Must(template.New("dockerfile").Parse(MkonionDockerfileTemplate))

func generateDockerfile(hasKey bool) (string, error) {
	config := new(bytes.Buffer)

	if err := dockerfileTemplate.Execute(config, struct {
		HasKey bool
	}{
		HasKey: hasKey,
	}); err != nil {
		return "", err
	}

	return config.String(), nil
}

func makeBuildContext(torrc []byte, privatekey []byte) (io.Reader, error) {
	hasKey := len(privatekey) > 0
	dockerfile, err := generateDockerfile(hasKey)
	if err != nil {
		return nil, err
	}

	files := []*FakeFile{{
		path: "torrc",
		mode: 0644,
		data: torrc,
	}, {
		path: "Dockerfile",
		mode: 0644,
		data: []byte(dockerfile),
	}}

	// XXX: This is probably slightly unsafe.
	if hasKey {
		files = append(files, &FakeFile{
			path: "private_key",
			mode: 0600,
			data: privatekey,
		})
	}

	return ArchiveContext(files)
}

func buildTorImage(cli *client.Client, ctx io.Reader) (string, error) {
	// XXX: There's currently no way to get the image ID of a build without
	//      manually parsing the output, or tagging the image. Since I'm not in
	//      the mood for the former, we can tag the build with a random name.
	//      Unfortunately, untagging of images isn't supported, so we'll have to
	//      use a name that allows us to not pollute the host.

	options := types.ImageBuildOptions{
		// XXX: If we SuppressOutput we can get just the image ID, but we lose
		//      being able to tell users what the status of the build is.
		//SuppressOutput: true,
		Tags:        []string{MkonionTag},
		Remove:      true,
		ForceRemove: true,
		Dockerfile:  "Dockerfile",
		Context:     ctx,
	}

	build, err := cli.ImageBuild(options)
	if err != nil {
		return "", err
	}

	// XXX: For some weird reason, at this point the build has not finished. We
	//      need to wait for build.Body to be closed. We might as well tell the
	//      user what the status of the build is.
	log.Infof("building %s", MkonionTag)
	dec := json.NewDecoder(build.Body)
	for {
		// Modified from pkg/jsonmessage in Docker.
		type JSONMessage struct {
			Stream string `json:"stream,omitempty"`
			Status string `json:"status,omitempty"`
		}

		// Decode the JSONMessages.
		var jm JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		jm.Stream = strings.TrimSpace(jm.Stream)
		jm.Status = strings.TrimSpace(jm.Status)

		// Log the status.
		if jm.Stream != "" {
			log.Info(jm.Stream)
		}
		if jm.Status != "" {
			log.Info(jm.Status)
		}
	}

	inspect, _, err := cli.ImageInspectWithRaw(MkonionTag, false)
	if err != nil {
		// XXX: Should probably clean up the built image here?
		return "", err
	}

	log.Infof("successfully built %s image", MkonionTag)
	return inspect.ID, nil
}

func runTorContainer(cli *client.Client, ident, imageID, network string) (string, error) {
	config := &types.ContainerCreateConfig{
		Name: ident,
		Config: &containerTypes.Config{
			Image: imageID,
		},
	}

	resp, err := cli.ContainerCreate(config.Config, config.HostConfig, config.NetworkingConfig, config.Name)
	if err != nil {
		return "", err
	}
	// TODO: Remove container on failure.

	for _, warning := range resp.Warnings {
		log.Warn(warning)
	}

	if err := cli.ContainerStart(resp.ID); err != nil {
		return "", err
	}

	// Connect to the network.
	if err := cli.NetworkConnect(network, resp.ID, nil); err != nil {
		return "", err
	}

	return resp.ID, err
}

type FakeBuildOptions struct {
	ident      string
	networkID  string
	torrc      []byte
	privatekey []byte
}

// FakeBuildRun builds and starts a new mkonion tor server container entirely
// in memory with no files created on the local machine.
func FakeBuildRun(cli *client.Client, options *FakeBuildOptions) (string, error) {
	ctx, err := makeBuildContext(options.torrc, options.privatekey)
	if err != nil {
		return "", fmt.Errorf("making build context: %s", err)
	}

	imageID, err := buildTorImage(cli, ctx)
	if err != nil {
		return "", fmt.Errorf("building image: %s", err)
	}

	containerID, err := runTorContainer(cli, options.ident, imageID, options.networkID)
	if err != nil {
		return "", fmt.Errorf("starting container: %s", err)
	}

	return containerID, nil
}
