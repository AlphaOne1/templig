// Copyright the templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the simple example.
// This example demonstrates the basic function of templig,
// namely the reading and deserialization of configuration files into an arbitrary data structure.
package main

import (
	"fmt"

	"github.com/AlphaOne1/templig"
)

// Config is the configuration structure that is to be filled by templig.
type Config struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
}

// main reads a single configuration file.
func main() {
	c, confErr := templig.FromFile[Config]("my_config.yaml")

	fmt.Printf("read errors: %v\n", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
	}
}
