// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig

import (
	"errors"
	"testing"
)

func TestEmptySource(t *testing.T) {
	t.Parallel()

	s := source{}

	if _, _, err := s.Reader(); !errors.Is(err, ErrNoConfigPaths) || !errors.Is(err, ErrNoConfigReaders) {
		t.Errorf("reading from empty source should have returned an error")
	}
}
