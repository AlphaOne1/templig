// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

// Package templig is the main package of the configuration library.
package templig

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"go.yaml.in/yaml/v4"
)

var (
	// ErrNoConfigReaders indicates that no configuration readers were provided to the function.
	ErrNoConfigReaders = errors.New("no configuration readers given")

	// ErrNoConfigPaths indicates that no configuration file paths were provided where at least one is required.
	ErrNoConfigPaths = errors.New("no configuration paths given")

	// ErrNoSecretRegexp indicates that no secret regular expression was provided where one is required.
	ErrNoSecretRegexp = errors.New("no secret regular expression given")
)

// Validator is the interface to facility validity checks on configuration types.
type Validator interface {
	// Validate is used to Validate a configuration.
	Validate() error
}

// Config is the generic structure holding the configuration information for the specified type.
type Config[T any] struct {
	node     *yaml.Node
	content  T
	secretRE *regexp.Regexp
}

// Get gives a pointer to the deserialized configuration. Get does not load the configuration anew and
// is principally inexpensive to call.
func (c *Config[T]) Get() *T {
	return &c.content
}

// overlay is called repeatedly and overlays the current intermediate configuration
// with the content of the given io.Reader.
func (c *Config[T]) overlay(r io.Reader) error {
	additionalConfig, aErr := fromSingle[yaml.Node](r)

	if aErr != nil {
		return aErr
	}

	if err := c.SetSecretRE(additionalConfig.secretRE); err != nil {
		return err
	}

	if c.node == nil {
		c.node = additionalConfig.Get()
	} else {
		merged, mergeErr := MergeYAMLNodes(c.node, additionalConfig.Get())

		if mergeErr != nil {
			return mergeErr
		}

		c.node = merged
	}

	return nil
}

// overlayFile opens a given configuration file and loads it as an intermediate using the overlay function.
func (c *Config[T]) overlayFile(path string) error {
	f, err := os.Open(filepath.Clean(path))

	if err != nil {
		return fmt.Errorf("could not open overlay file %v: %w", path, err)
	}

	defer func() { _ = f.Close() }()

	return c.overlay(f)
}

// fromSingle reads a configuration from the single given io.Reader and
// runs - if necessary - the contained template functions.
func fromSingle[T any](r io.Reader) (*Config[T], error) {
	var config Config[T]

	if err := config.SetSecretRE(SecretRE); err != nil {
		return nil, err
	}

	fileContent, err := io.ReadAll(r)

	if err != nil {
		return nil, fmt.Errorf("could not read from reader: %w", err)
	}

	var tmpl *template.Template

	if tmpl, err = template.
		New("config").
		Funcs(templigFunctions()).
		Parse(string(fileContent)); err != nil {
		return nil, fmt.Errorf("could not parse template: %w", err)
	}

	var b bytes.Buffer

	if err = tmpl.Execute(&b, nil); err != nil {
		return nil, fmt.Errorf("could not execute template: %w", err)
	}

	if decodeErr := yaml.NewDecoder(&b).Decode(&config.content); decodeErr != nil {
		return nil, fmt.Errorf("could not parse configuration: %w", decodeErr)
	}

	return &config, nil
}

// Validate checks if the configuration is valid if the content fulfills the Validator interface.
func (c *Config[T]) Validate() error {
	if v, ok := any(&c.content).(Validator); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}

// From reads a configuration from the given set of io.Reader.
func From[T any](readers ...io.Reader) (*Config[T], error) {
	if len(readers) == 0 {
		return nil, ErrNoConfigReaders
	}

	var config *Config[T]
	var decodeErr error
	var validateErr error

	if len(readers) == 1 {
		// to optimize the most common case of a single reader, we do not need to
		// go over the yaml.Node structure first.
		config, decodeErr = fromSingle[T](readers[0])
	} else {
		config = new(Config[T])

		for _, v := range readers {
			if err := config.overlay(v); err != nil {
				return nil, err
			}
		}

		decodeErr = config.node.Decode(&config.content)

		// cleanup
		config.node = nil
	}

	if decodeErr == nil {
		validateErr = config.Validate()
	}

	if resultErr := errors.Join(decodeErr, validateErr); resultErr != nil {
		return nil, resultErr
	}

	return config, nil
}

// To writes a configuration to the given io.Writer.
func (c *Config[T]) To(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	err := wrapError("could not encode configuration", enc.Encode(&c.content))
	encCloseErr := enc.Close()

	return errors.Join(err, encCloseErr)
}

// ToSecretsHidden writes the configuration to the given io.Writer and hides secret values using the [SecretRE].
// Strings are replaced with the number of * corresponding to their length.
// Substructures containing secrets are replaced with a single '*'.
// The following example
//
//	id: id0
//	secrets:
//	  - secret0
//	  - secret1
//
// thus will be replaced by
//
//	id: id0
//	secrets: *
func (c *Config[T]) ToSecretsHidden(w io.Writer) error {
	var writeErr error
	var encCloseErr error
	node := yaml.Node{}

	encodeErr := node.Encode(c.content)

	if encodeErr == nil {
		HideSecrets(&node, true, c.secretRE)
		enc := yaml.NewEncoder(w)
		writeErr = enc.Encode(node)
		encCloseErr = enc.Close()
	}

	return errors.Join(encodeErr, writeErr, encCloseErr)
}

// ToSecretsHiddenStructured writes the configuration to the given io.Writer
// and hides secret values using the [SecretRE].
// Strings are replaced with the number of * corresponding to their length.
// Substructures containing secrets are replaced with a corresponding structure of '*'.
// The following example
//
//	id: id0
//	secrets:
//	  - secret0
//	  - secret1
//
// thus will be replaced by
//
//	id: id0
//	secrets:
//	  - *******
//	  - *******
func (c *Config[T]) ToSecretsHiddenStructured(w io.Writer) error {
	var writeErr error
	var encCloseErr error
	node := yaml.Node{}

	encodeErr := node.Encode(c.content)

	if encodeErr == nil {
		HideSecrets(&node, false, c.secretRE)
		enc := yaml.NewEncoder(w)
		writeErr = enc.Encode(node)
		encCloseErr = enc.Close()
	}

	return errors.Join(encodeErr, writeErr, encCloseErr)
}

// FromFile loads a series of configuration files. The first file is considered the base, all others are
// loaded on top of that one using the [MergeYAMLNodes] functionality.
func FromFile[T any](paths ...string) (*Config[T], error) {
	if len(paths) == 0 {
		return nil, ErrNoConfigPaths
	}

	config := new(Config[T])
	var decodeErr error
	var validateErr error

	if len(paths) == 1 {
		// to optimize the most common case of a single file, we do not need to
		// go over the yaml.Node structure first.
		f, err := os.Open(paths[0])

		if err != nil {
			return nil, fmt.Errorf("could not open %s: %w", paths[0], err)
		}

		defer func() { _ = f.Close() }()

		config, decodeErr = fromSingle[T](f)
	} else {
		for _, addOn := range paths {
			aErr := config.overlayFile(addOn)

			if aErr != nil {
				return nil, aErr
			}
		}

		decodeErr = config.node.Decode(&config.content)

		// cleanup
		config.node = nil
	}

	if decodeErr == nil {
		validateErr = config.Validate()
	}

	if resultErr := errors.Join(decodeErr, validateErr); resultErr != nil {
		return nil, resultErr
	}

	return config, nil
}

// FromFiles loads a series of configuration files. The first file is considered the base, all others are
// loaded on top of that one using the [MergeYAMLNodes] functionality.
//
// Deprecated: As of version 'v0.6.0' this function is deprecated and will be removed in the next major release.
func FromFiles[T any](paths []string) (*Config[T], error) {
	return FromFile[T](paths...)
}

// ToFile saves a configuration to a file with the given name, replacing it in case.
func (c *Config[T]) ToFile(path string) error {
	f, err := os.Create(filepath.Clean(path))

	if err != nil {
		return fmt.Errorf("could not create file %s: %w", path, err)
	}

	defer func() { _ = f.Close() }()

	return c.To(f)
}

// SecretRE returns a copy of the regular expression used for hiding secrets of that specific instance.
func (c *Config[T]) SecretRE() *regexp.Regexp {
	tmp := *c.secretRE

	return &tmp
}

// SetSecretRE sets the regular expression to be used for hiding secrets of that specific instance. It must not be nil.
func (c *Config[T]) SetSecretRE(re *regexp.Regexp) error {
	if re == nil {
		return ErrNoSecretRegexp
	}

	c.secretRE = re

	return nil
}
