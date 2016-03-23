// mkonion: create a Tor onion service for existing Docker containers
// Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"math/rand"
	"time"
)

const (
	identifierPrefix = "mkonion_"
	identifierLength = 16
)

func init() {
	// We need to seed the randomness.
	now := time.Now()
	rand.Seed(now.Unix() * int64(now.Nanosecond()))
}

func generateIdentifier() string {
	const (
		identLetters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numLetters   = len(identLetters)
	)

	var ident string
	for i := 0; i < identifierLength; i++ {
		ident += string(identLetters[rand.Intn(numLetters)])
	}

	return identifierPrefix + ident
}
