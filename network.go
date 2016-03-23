// mkonion: create a Tor onion service for existing Docker containers
// Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
