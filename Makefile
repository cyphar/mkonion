# mkonion: create a Tor onion service for existing Docker containers
# Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

DOCKER=docker
GO=go

SRC=config.go fakebuild.go fakefile.go flag.go main.go name.go network.go hostname.go
OUT=bin

.PHONY: docker

docker: $(SRC)
	@mkdir -p $(OUT)
	$(DOCKER) build -t mkonion/build:dev .
	$(DOCKER) run -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) \
		-v $(PWD)/$(OUT):/go/src/github.com/cyphar/mkonion/$(OUT) mkonion/build:dev make OUT=$(OUT) mkonion

mkonion: $(SRC)
	@mkdir -p $(OUT)
	CGO_ENABLED=0 $(GO) build -a -installsuffix cgo -ldflags '-s' -o $(OUT)/mkonion $(SRC)
