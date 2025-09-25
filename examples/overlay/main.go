// SPDX-FileCopyrightText: 2025 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the overlay example.
// This example demonstrates the basic overlay function of templig.
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

// main reads a configuration file with an overlay.
func main() {
	c, confErr := templig.FromFile[Config](
		"my_config.yaml",
		"my_prod_overlay.yaml",
	)

	fmt.Printf("read errors: %v\n", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
	}
}
