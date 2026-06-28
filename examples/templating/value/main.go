// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the templating variable `.Value` example.
// This example demonstrates the use of the `.Value` variable in a templated configuration.
// The use of the `required` function is demonstrated in conjunction with `.Value`.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlphaOne1/templig"
)

// Config is the configuration structure that is to be filled by templig.
type Config struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
	Pass string `yaml:"pass"`
}

// main reads a configuration file. The configuration file then uses the .Value variable to read the password.
// To insert a custom variable into the configuration, the WithValue option is used.
func main() {
	c, confErr := templig.New[Config](
		templig.WithFile[Config]("my_config.yaml"),
		templig.WithValue[Config]("pass", "secret"))

	fmt.Printf("read errors: %v\n", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
		fmt.Printf("Pass: %v\n", strings.Repeat("*", len(c.Get().Pass)))
		fmt.Println("Config printed by templig with hidden secrets:")

		_ = c.ToSecretsHiddenStructured(os.Stdout)
	}
}
