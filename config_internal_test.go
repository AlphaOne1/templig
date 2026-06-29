// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig

import (
	"errors"
	"testing"
)

func TestEmptySource0(t *testing.T) {
	t.Parallel()

	s := source{}

	_, err := s.Reader()

	if !errors.Is(err, ErrNoConfigPaths) || !errors.Is(err, ErrNoConfigReaders) {
		t.Errorf("reading from empty source should have returned an error")
	}
}

func TestNoSources(t *testing.T) {
	t.Parallel()

	c := Config[int]{}

	if err := c.addSources(); !errors.Is(err, ErrNoConfigPaths) || !errors.Is(err, ErrNoConfigReaders) {
		t.Errorf("adding no sources should have returned an error")
	}
}
