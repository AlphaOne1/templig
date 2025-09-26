// SPDX-FileCopyrightText: 2025 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"testing"
)

func TestMainGood(t *testing.T) {
	t.Parallel()

	os.Args = []string{"main"}
	main()
}
