// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig

import (
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// TemplateFunctions is a template.FuncMap that allows to globally remove the additional templig template functions or
// even add own functions on top of what is already provided. For user-provided functions, please use the prefix `uP`,
// that is guaranteed to never be used as a templig provided function. Access to this variable
// is not synchronized, thus modifying it shall be done before working with templig.
var TemplateFunctions = template.FuncMap{ //nolint:gochecknoglobals
	"arg":      argumentValue,
	"hasArg":   argumentPresent,
	"required": required,
	"read":     readFile,
}

// templigFunctions gives all the functions that are enabled for the templating engine.
func templigFunctions() template.FuncMap {
	result := sprig.TxtFuncMap()

	maps.Insert(result, maps.All(TemplateFunctions))

	return result
}

// required is a template function to indicate that the second argument cannot be empty or nil.
func required(warn string, val any) (any, error) {
	if s, ok := val.(string); val == nil || (ok && s == "") {
		return val, errors.New(warn) //nolint:err113
	}

	return val, nil
}

// readFile is a template function to read a file and store its content into a string.
// If the file does not exist, an empty string is generated, facilitating the use of `required` for customized
// user interaction.
func readFile(fileName string) (any, error) {
	file, err := os.Open(filepath.Clean(fileName))

	if err != nil {
		return "", nil
	}

	defer func() { _ = file.Close() }()

	content, readErr := io.ReadAll(file)

	return string(content), readErr
}

func argumentIndexMatch(name string) func(string) bool {
	return func(s string) bool {
		tmp := strings.TrimLeft(s, "-")

		return len(tmp) != len(s) &&
			(tmp == name ||
				len(tmp) > len(name) && strings.HasPrefix(tmp, name) && tmp[len(name)] == '=')
	}
}

func argumentValue(name string) (any, error) {
	index := slices.IndexFunc(os.Args, argumentIndexMatch(name))

	// handle arguments that give the value using assignment
	if index >= 0 && strings.Contains(os.Args[index], "=") {
		argument := strings.SplitN(os.Args[index], "=", 2)

		if len(argument) == 2 {
			return argument[1], nil
		}
	}

	// handle arguments with the value in the next argument
	// (that then may not start with a dash)
	if index >= 0 &&
		len(os.Args) > index+1 &&
		!strings.HasPrefix(os.Args[index+1], "-") {

		return os.Args[index+1], nil
	}

	// no argument value given
	return "", nil
}

func argumentPresent(name string) (any, error) {
	index := slices.IndexFunc(os.Args, argumentIndexMatch(name))

	if index >= 0 {
		return true, nil
	}

	return false, nil
}
