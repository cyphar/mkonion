// mkonion: create a Tor onion service for existing Docker containers
// Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"strings"
)

type flagList []string

func (fl *flagList) String() string {
	return "[" + strings.Join(*fl, ", ") + "]"
}

func (fl *flagList) Set(field string) error {
	*fl = append(*fl, field)
	return nil
}
