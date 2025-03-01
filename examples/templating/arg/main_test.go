// Copyright the templig contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"testing"
)

func TestMainGood(t *testing.T) {
	os.Args = []string{"main", "--pass=bogus"}
	main()
}

func TestMainBad(t *testing.T) {
	os.Args = []string{"main"}
	main()
}
