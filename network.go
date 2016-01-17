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
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	networkTypes "github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
)

// CreateOnionNetwork creates a new bridge network with a random (but recognisable)
// name. If it can't create a name after XXX attempts, it will return an error.
func CreateOnionNetwork(cli *client.Client, ident string) (string, error) {
	options := types.NetworkCreate{
		Name:           ident,
		CheckDuplicate: true,
		Driver:         "bridge",
	}

	resp, err := cli.NetworkCreate(options)
	if err != nil {
		// TODO: Retry if we get "already exists".
		return "", err
	}

	if resp.Warning != "" {
		log.Warn(resp.Warning)
	}

	return ident, nil
}

// ConnectOnionNetwork connects a target container to the onion network, allowing
// the container to be accessed by the Tor relay container.
func ConnectOnionNetwork(cli *client.Client, target, network string) error {
	// XXX: Should configure this to use a subnet like 10.x.x.x.
	options := &networkTypes.EndpointSettings{}
	return cli.NetworkConnect(network, target, options)
}

// PurgeOnionNetwork purges an onion network, disconnecting all containers with
// it. We assume that nobody is adding containers to this network.
func PurgeOnionNetwork(cli *client.Client, network string) error {
	inspect, err := cli.NetworkInspect(network)
	if err != nil {
		return err
	}

	for container, _ := range inspect.Containers {
		log.Infof("purge network %s: disconnecting container %s", network, container)
		if err := cli.NetworkDisconnect(network, container, true); err != nil {
			return err
		}
	}

	return cli.NetworkRemove(network)
}

// FindTargetPorts finds the set of ports EXPOSE'd on the target container. This
// includes non-TCP ports, so callers should make sure they exclude protocols
// not supported by Tor.
func FindTargetPorts(cli *client.Client, target string) ([]nat.Port, error) {
	inspect, err := cli.ContainerInspect(target)
	if err != nil {
		return nil, err
	}

	// Make sure we don't dereference nils.
	if inspect.NetworkSettings == nil || inspect.NetworkSettings.Ports == nil {
		return nil, fmt.Errorf("inspect container: network settings not available")
	}

	// Get keys from map.
	var ports []nat.Port
	for port, _ := range inspect.NetworkSettings.Ports {
		ports = append(ports, port)
	}

	return ports, nil
}

// FindOnionIPAddress finds the IP address of a target container that is connected
// to the given network. This IP address is accessible from any other container
// connected to the same network.
func FindOnionIPAddress(cli *client.Client, target, network string) (string, error) {
	inspect, err := cli.ContainerInspect(target)
	if err != nil {
		return "", err
	}

	endpoint, ok := inspect.NetworkSettings.Networks[network]
	if !ok {
		return "", fmt.Errorf("inspect container: container '%s' not connected to network '%s'", target, network)
	}

	return endpoint.IPAddress, nil
}
