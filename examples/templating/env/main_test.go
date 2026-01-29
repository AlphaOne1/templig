// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"testing"
)

func TestMainGood(t *testing.T) {
	os.Args = []string{"main"}
	t.Setenv("PASSWORD", "bogus")
	main()
}

func TestMainBad(t *testing.T) {
	t.Parallel()

	os.Args = []string{"main"}
	main()
}
