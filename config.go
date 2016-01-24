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
HiddenServiceDir /var/run/tor/hidden_service
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
