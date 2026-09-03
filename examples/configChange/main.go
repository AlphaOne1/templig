// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the configChange example.
//
// It illustrates how to use templig to detect changes in a configuration file. templig cannot provide appropriate
// change management functionality, as the sources of the configuration data vary and are not controlled by templig
// itself. E.g., a network connection could be given to templig.WithReader, that cannot be easily read again in a
// comparison run.
//
// This example uses a hash function instead of e.g., reflect.DeepEqual. That way, a potentially large configuration
// structure does not have to be preserved for a prolonged time in memory.
package main

import (
	"bytes"
	"crypto/sha3"
	"log/slog"

	"github.com/AlphaOne1/templig"
)

// Config is a deliberately simple structure.
type Config struct {
	Val int `yaml:"val"`
}

// main reads a configuration stream and creates its hash value. After a second stream is read, and its hash value is
// compared to the first one.
func main() {
	buf := bytes.NewBufferString("val: 2")
	c, _ := templig.New[Config](templig.WithReader(buf))

	hasher := sha3.New384() // create a new hasher
	_ = c.To(hasher)        // fill it with the unmodified configuration (no password hiding!)

	oldHash := hasher.Sum(nil) // get the finished hash value

	buf = bytes.NewBufferString("val: 3")
	cNew, _ := templig.New[Config](templig.WithReader(buf))

	hasher.Reset()      // reset the previous hasher for new use
	_ = cNew.To(hasher) // fill it with the unmodified content of the configuration to compare to

	newHash := hasher.Sum(nil) // get the finished second hash value

	if !bytes.Equal(oldHash, newHash) {
		slog.Info("config changed")
	}
}
