// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the templating function `arg` example.
// This example demonstrates the use of the `arg` functions in a templated configuration.
// The use of the `required` function is demonstrated in conjunction with `arg`.
package main

import (
	"fmt"
	"strings"

	"github.com/AlphaOne1/templig"
)

// Config is the configuration structure that is to be filled by templig.
type Config struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
	Pass string `yaml:"pass"`
}

// main reads one configuration file. This configuration file uses then the templig functions.
// Note that this main function does not have any specific adjustments and is basically identical to the simple case.
func main() {
	c, confErr := templig.FromFile[Config]("my_config.yaml")

	fmt.Printf("read errors: %v\n", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
		fmt.Printf("Pass: %v\n", strings.Repeat("*", len(c.Get().Pass)))
	}
}
