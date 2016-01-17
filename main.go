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
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
)

func mkonion() (err error) {
	flag.Parse()
	oTargetContainer := flag.Arg(0)

	if flag.NArg() != 1 || oTargetContainer == "" {
		flag.Usage()
		return fmt.Errorf("must specify a container to create an onion service for")
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return fmt.Errorf("connecting to client: %s", err)
	}

	ident := generateIdentifier()
	networkID, err := CreateOnionNetwork(cli, ident)
	if err != nil {
		return fmt.Errorf("creating onion network: %s", err)
	}
	log.WithFields(log.Fields{
		"network": ident,
	}).Info("created onion network")
	defer func() {
		if err != nil {
			if err := PurgeOnionNetwork(cli, networkID); err != nil {
				log.Warnf("purge onion network: %s", err)
			}
		}
	}()

	if err := ConnectOnionNetwork(cli, oTargetContainer, networkID); err != nil {
		return fmt.Errorf("connecting target to onion network: %s", err)
	}
	log.WithFields(log.Fields{
		"network":   ident,
		"container": oTargetContainer,
	}).Info("attached container to onion network")

	ip, err := FindOnionIPAddress(cli, oTargetContainer, networkID)
	if err != nil {
		return fmt.Errorf("finding target onion ip: %s", err)
	}
	log.WithFields(log.Fields{
		"network":   ident,
		"container": oTargetContainer,
		"ip":        ip,
	}).Info("found target address")

	ports, err := FindTargetPorts(cli, oTargetContainer)
	if err != nil {
		return fmt.Errorf("finding target ports: %s", err)
	}

	for _, port := range ports {
		log.Infof("forwarding port: %s", port)
	}

	torrc, err := GenerateConfig(cli, ip, ports)
	if err != nil {
		return fmt.Errorf("generating torrc: %s", err)
	}
	log.Info("generated torrc config")

	buildOptions := &FakeBuildOptions{
		ident:     ident,
		networkID: networkID,
		torrc:     torrc,
	}

	containerID, err := FakeBuildRun(cli, buildOptions)
	if err != nil {
		return fmt.Errorf("starting tor daemon: %s", err)
	}
	log.WithFields(log.Fields{
		"container": containerID,
	}).Infof("tor daemon started")

	return nil
}

func main() {
	if err := mkonion(); err != nil {
		log.Fatal(err)
	}
}
