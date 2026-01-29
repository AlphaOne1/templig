// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the validation example.
// This example demonstrates the use of the validation functionality.
package main

import (
	"errors"
	"fmt"

	"github.com/AlphaOne1/templig"
)

// Config is the configuration structure that is to be filled by templig.
type Config struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
}

// Validate fulfills the Validator interface provided by templig.
// This method is called if it is defined. It influences the outcome of the configuration reading.
func (c *Config) Validate() error {
	var result []error

	if len(c.Name) == 0 {
		result = append(result, errors.New("name is required"))
	}

	if c.ID < 0 {
		result = append(result, errors.New("id greater than zero is required"))
	}

	return errors.Join(result...)
}

// main reads a configuration file. Note that the validation is done inside the template reading, so the
// code in main does not need specific modifications.
func main() {
	_, confErr := templig.FromFile[Config]("my_config_bad.yaml")

	// read errors bad config: name is required
	fmt.Printf("read errors bad config: %v\n", confErr)

	c, confErr := templig.FromFile[Config]("my_config_good.yaml")

	fmt.Printf("read errors good config: %v\n", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
	}
}
