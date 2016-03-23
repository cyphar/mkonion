# mkonion: create a Tor onion service for existing Docker containers
# Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

FROM golang:1.5.3
MAINTAINER "Aleksa Sarai <cyphar@cyphar.com>"

RUN go get -v github.com/tools/godep

WORKDIR /go/src/github.com/cyphar/mkonion
COPY . /go/src/github.com/cyphar/mkonion
RUN godep restore
