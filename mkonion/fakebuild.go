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
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	containerTypes "github.com/docker/engine-api/types/container"
	networkTypes "github.com/docker/engine-api/types/network"
)

// Building a Docker image usually requires a real filesystem in order to create
// the tar archive sent to the Docker daemon. But if all you have a single
// configuration file, it's a pain to do that. In a sense, it's much nicer to
// embed all of the required information into the `mkonion` binary so we don't
// have to touch the filesystem.

const (
	MkonionTag        = "mkonion/tor:latest"
	MkonionDockerfile = `
	FROM ubuntu:14.04
	`
)

func makeBuildContext(torrc []byte) (io.Reader, error) {
	return ArchiveContext([]*FakeFile{{
		Path: "Dockerfile",
		Data: []byte(MkonionDockerfile),
	}, {
		Path: "torrc",
		Data: torrc,
	}})
}

func buildTorImage(cli *client.Client, ctx io.Reader) (string, error) {
	// XXX: There's currently no way to get the image ID of a build without
	//      manually parsing the output, or tagging the image. Since I'm not in
	//      the mood for the former, we can tag the build with a random name.
	//      Unfortunately, untagging of images isn't supported, so we'll have to
	//      use a name that allows us to not pollute the host.

	options := types.ImageBuildOptions{
		//SuppressOutput: true,
		Tags:        []string{MkonionTag},
		Remove:      true,
		ForceRemove: true,
		Dockerfile:  "Dockerfile",
		Context:     ctx,
	}

	_, err := cli.ImageBuild(options)
	if err != nil {
		return "", err
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
		NetworkingConfig: &networkTypes.NetworkingConfig{
			// Connect to the network.
			EndpointsConfig: map[string]*networkTypes.EndpointSettings{
				network: nil,
			},
		},
	}

	resp, err := cli.ContainerCreate(config.Config, config.HostConfig, config.NetworkingConfig, config.Name)
	if err != nil {
		return "", err
	}

	for _, warning := range resp.Warnings {
		log.Warn(warning)
	}

	if err := cli.ContainerStart(resp.ID); err != nil {
		return "", err
	}

	return resp.ID, err
}

type FakeBuildOptions struct {
	ident     string
	networkID string
	torrc     []byte
}

// FakeBuildRun builds and starts a new mkonion tor server container entirely
// in memory with no files created on the local machine.
func FakeBuildRun(cli *client.Client, options *FakeBuildOptions) (string, error) {
	ctx, err := makeBuildContext(options.torrc)
	if err != nil {
		return "", err
	}

	imageID, err := buildTorImage(cli, ctx)
	if err != nil {
		return "", err
	}

	return runTorContainer(cli, options.ident, imageID, options.networkID)
}
