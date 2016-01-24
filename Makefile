# mkonion: create a Tor onion service for existing Docker containers
# Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

# Permission is hereby granted, free of charge, to any person obtaining a copy of
# this software and associated documentation files (the "Software"), to deal in
# the Software without restriction, including without limitation the rights to
# use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:

# 1. The above copyright notice and this permission notice shall be included in
#    all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

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
