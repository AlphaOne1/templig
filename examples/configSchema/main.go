// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main of the configSchema example.
package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/AlphaOne1/templig"
	"github.com/xeipuuv/gojsonschema"
)

// Validate uses the xeipuuv/gojsonschemna library to load a json schema
// and validates the loaded configuration against that schema.
// There are of course other libraries, like the atombender/go-jsonschema, but they do not support to validate the
// configuration object directly but need a prior conversion to map[string]any or JSON. It is important to run the
// validation on the final configuration, as there may be overlays and template invocations that influence the
// validity.
func (c *Config) Validate() error {
	schemaLoader := gojsonschema.NewReferenceLoader("file://./schema.json")
	documentLoader := gojsonschema.NewGoLoader(c)

	result, resultErr := gojsonschema.Validate(schemaLoader, documentLoader)

	if resultErr != nil {
		return fmt.Errorf("could not validate: %w", resultErr)
	}

	if !result.Valid() {
		return fmt.Errorf("%v", result)
	}

	return nil
}

// main reads the configuration and prints out the final configuration hiding the contained secrets.
func main() {
	config, configErr := templig.FromFile[Config]("config.yaml")

	if configErr != nil {
		slog.Error("could not read configuration", slog.String("error", configErr.Error()))
		os.Exit(-1)
	}

	slog.Debug("configuration loaded successfully")

	fmt.Println("configuration:")
	_ = config.ToSecretsHiddenStructured(os.Stdout)
}
