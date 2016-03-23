// mkonion: create a Tor onion service for existing Docker containers
// Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"text/template"

	"github.com/docker/engine-api/client"
)

const (
	// Describes the entire torrc template.
	// TODO: Make the hidden_service path customisable.
	torTemplate = `
# Disable SOCKS, we're only running as a hidden service.
SocksPort 0

# Set up hidden service.
HiddenServiceDir /var/lib/tor/hidden_service
{{range .Targets}}
HiddenServicePort {{.ExternalPort}} {{.}}
{{end}}
`
)

type TargetIP struct {
	Addr         string
	InternalPort string
	ExternalPort string
}

func (t TargetIP) String() string {
	return t.Addr + ":" + t.InternalPort
}

// XXX: This is probably very horrible.
var confTemplate = template.Must(template.New("tor").Parse(torTemplate))

func GenerateTargetMappings(ip string, mappings map[string]string) []TargetIP {
	var targets []TargetIP
	for external, internal := range mappings {
		targets = append(targets, TargetIP{
			Addr:         ip,
			InternalPort: internal,
			ExternalPort: external,
		})
	}
	return targets
}

// GenerateConfig generates a configuraton file for a target container for a
// given network. This is returned as a string, and warnings are logged if there
// are any non-TCP ports exposed on the container.
func GenerateConfig(cli *client.Client, targets []TargetIP) ([]byte, error) {
	config := new(bytes.Buffer)

	if err := confTemplate.Execute(config, struct {
		Targets []TargetIP
	}{
		Targets: targets,
	}); err != nil {
		return nil, err
	}

	return config.Bytes(), nil
}
