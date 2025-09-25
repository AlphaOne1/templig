// SPDX-FileCopyrightText: 2025 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the templating function `read` example.
// This example demonstrates the use of the `read` functions in a templated configuration.
// The use of the `required` function is demonstrated in conjunction with `read`.
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

// main reads a configuration file. That configuration file then uses the read templig function to read a file with
// additional information. Note that there is no specific adaption in this main function, to make this possible.
func main() {
	c, confErr := templig.FromFile[Config]("my_config.yaml")

	fmt.Printf("read errors: %v\n", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
		fmt.Printf("Pass: %v\n", strings.Repeat("*", len(c.Get().Pass)))
	}
}
