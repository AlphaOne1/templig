// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the templating function `hasArg` example.
// This example demonstrates the use of the `hasArg` functions in a templated configuration.
// It is used in conjunction with `required` and `env`.
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

// main reads a configuration file. The configuration file then uses the hasArg templig function to see, if the
// parameter passEnv has been set, to decide, if it should be read from the environment or the default be used.
// Note that there is no specific adaption in this main function to get this functionality.
func main() {
	c, confErr := templig.FromFile[Config]("my_config.yaml")

	fmt.Printf("read errors: %v\n", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
		fmt.Printf("Pass: %v\n", strings.Repeat("*", len(c.Get().Pass)))
	}
}
