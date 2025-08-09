<!-- markdownlint-disable MD013 MD033 MD041 -->
<p align="center">
    <img src="templig_logo.svg" width="25%" alt="Logo"><br>
    <a href="https://github.com/AlphaOne1/templig/actions/workflows/test.yml"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/templig/actions/workflows/test.yml/badge.svg"
             alt="Test Pipeline Result">
    </a>
    <a href="https://github.com/AlphaOne1/templig/actions/workflows/codeql.yml"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/templig/actions/workflows/codeql.yml/badge.svg"
             alt="CodeQL Pipeline Result">
    </a>
    <a href="https://github.com/AlphaOne1/templig/actions/workflows/security.yml"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/templig/actions/workflows/security.yml/badge.svg"
             alt="Security Pipeline Result">
    </a>
    <a href="https://goreportcard.com/report/github.com/AlphaOne1/templig"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://goreportcard.com/badge/github.com/AlphaOne1/templig"
             alt="Go Report Card">
    </a>
    <a href="https://codecov.io/github/AlphaOne1/templig"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://codecov.io/github/AlphaOne1/templig/graph/badge.svg?token=P18EOCUPU8"
             alt="Code Coverage">
    </a>
    <img src="https://img.shields.io/coderabbit/prs/github/AlphaOne1/templig"
         alt="CodeRabbit Reviews">
    <a href="https://www.bestpractices.dev/projects/9789"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://www.bestpractices.dev/projects/9789/badge"
             alt="OpenSSF Best Practices">
    </a>
    <a href="https://scorecard.dev/viewer/?uri=github.com/AlphaOne1/templig"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://api.scorecard.dev/projects/github.com/AlphaOne1/templig/badge"
             alt="OpenSSF Scorecard">
    </a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2FAlphaOne1%2Ftemplig?ref=badge_shield&issueType=license"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2FAlphaOne1%2Ftemplig.svg?type=shield&issueType=license"
            alt="FOSSA License Status">
    </a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2FAlphaOne1%2Ftemplig?ref=badge_shield&issueType=security"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2FAlphaOne1%2Ftemplig.svg?type=shield&issueType=security"
             alt="FOSSA Security Status">
    </a>
    <a href="https://godoc.org/github.com/AlphaOne1/templig"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://godoc.org/github.com/AlphaOne1/templig?status.svg"
             alt="GoDoc Reference">
    </a>
</p>
<!-- markdownlint-enable MD013 MD033 MD041 -->

templig
=======

*templig* (pronounced [ˈtɛmplɪç]) is a non-intrusive configuration library that utilizes the text-templating
engine of Go and the functions best known from [Helm](https://github.com/helm/helm) charts, originating from
[Masterminds/sprig](https://github.com/Masterminds/sprig).

Its primary goal is to enable dynamic configuration files, that have access to the system environment to fill
information using functions like `env` and `read`. To facilitate different environments, overlays can be defined
that amend a base configuration with environment-specific attributes and changes.
Configurations that implement the `Validator` interface also have automated checking enabled upon loading.

This is not the first configuration library and surely will not be the last. There exist alternatives, the most
elaborate of them may be [viper](https://github.com/spf13/viper). The difference to basically all of these is that
they burden the developer to provide all the means to gather the configuration information. So if the developer
does not foresee a means to read a value from the environment, a user cannot use this. *templig* turns that
around and gives the developer a simple interface to do what he wants—read a config—and gives the user the means
to compile his configuration in whatever way he sees fit. Experience shows that the target system environments can
be extremely diverse, and limiting the possibilities of end users directly limits the spectrum of application.

Installation
------------

To install *templig*, you can use the following command:

```bash
$ go get github.com/AlphaOne1/templig
```


Getting Started
---------------

### Simple Case

Having a configuration file like the following:

```yaml
id:   23
name: Interesting Name
```

The code to read that file would look like this:

```go
package main

import (
	"fmt"

	"github.com/AlphaOne1/templig"
)

// Config is the configuration structure
type Config struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
}

// main will read and display the configuration
func main() {
	c, confErr := templig.FromFile[Config]("my_config.yaml")

	fmt.Printf("read errors: %v", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
	}
}
```

The `Get` method gives a pointer to the internally held Config structure that the user supplied. The pointer is always
non-nil, so additional nil-checks are not necessary. Running that program would give:

```text
read errors: <nil>
ID:   23
Name: Interesting Name
```


### Reading with Overlays

Having a base configuration file `my_config.yaml` like the following:

```yaml
id:   23
name: Interesting DevName
```

and a file that contains specific configuration for e.g. the production environment `my_prod_overlay.yaml`:

```yaml
name: Important ProdName
```

The code to read those files would look like this:

```go
package main

import (
	"fmt"

	"github.com/AlphaOne1/templig"
)

// Config is the configuration structure
type Config struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
}

// main will read and display the configuration
func main() {
	c, confErr := templig.FromFile[Config](
		"my_config.yaml",
		"my_prod_overlay.yaml",
	)

	fmt.Printf("read errors: %v", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
	}
}
```

That way, the different configuration files are read in order, with the first one as the base. Every additional file
gives changes to all the ones read before. In this example, changing the name. Running this program would give:

```text
read errors: <nil>
ID:   23
Name: Important ProdName
```

As expected, the value of `Name` was replaced by the one provided in overlay configuration.


### Template Functionality
#### Overview

*templig* supports templating the configuration files. In addition to the basic templating functions provided by the Go
`text/template` library, *templig* includes the functions from [sprig](http://github.com/Masterminds/sprig), which are perhaps best known for
their use in [Helm](https://github.com/helm/helm) charts. On top of that, the following functions are provided for
convenience:

| Function | Description                                                         | Example                            |
|----------|---------------------------------------------------------------------|------------------------------------|
| arg      | reads the value of the command line argument with the given name    | [Link](examples/templating/arg)    |
| hasArg   | true if an argument with the given name is present, false otherwise | [Link](examples/templating/hasArg) |
| required | checks that its second argument is not zero length or nil           | [Link](examples/templating/env)    |
| read     | reads the content of a file                                         | [Link](examples/templating/read)   |

The expansion of the templated parts is done __before__ overlaying takes place. Any errors of templating will thus be
displayed in their respective source locations.


#### Reading Environment

Having a templated configuration file like this one:

```yaml
id:   23
name: Interesting Name
pass: {{ env "PASSWORD" | required "password required" | quote }}
```

or this one:

```yaml
id:   23
name: Interesting Name
pass: {{ read "pass.txt" | required "password required" | quote }}
```

One can see the templating code between the double curly braces `{{` and `}}`.
The following program is essentially the same as in the [Simple Case](#simple-case).
It just adds the `pass` field to the configuration:

```go
package main

import (
	"fmt"
	"strings"

	"github.com/AlphaOne1/templig"
)

// Config is the configuration structure
type Config struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
	Pass string `yaml:"pass"`
}

// main will read and display the configuration
func main() {
	c, confErr := templig.FromFile[Config]("my_config.yaml")

	fmt.Printf("read errors: %v", confErr)

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
		fmt.Printf("Pass: %v\n", strings.Repeat("*", len(c.Get().Pass)))
	}
}
```


### Validation

The templating facilities allow also for a wide range of tests, but depend on the configuration file read. As it is
most likely user supplied, possible consistency checks are not reliable in the form of template code.
For this purpose, *templig* also allows for the configuration structure to implement the `Validator` interface.
Implementing types provide a function `Validate` that allows *templig* to check __after__ the configuration was read,
if its structure should be considered valid and report errors accordingly.

```go
package main

import (
    "errors"
    "fmt"

	"github.com/AlphaOne1/templig"
)

// Config is the configuration structure
type Config struct {
    ID   int    `yaml:"id"`
    Name string `yaml:"name"`
}

// Validate fulfills the Validator interface provided by templig.
// This method is called, if it is defined. It influences the outcome of the configuration reading.
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

// main will read and display the configuration
func main() {
	c, confErr := templig.FromFile[Config]("my_config_good.yaml")

	if confErr == nil {
		fmt.Printf("ID:   %v\n", c.Get().ID)
		fmt.Printf("Name: %v\n", c.Get().Name)
	}
}
```

Validation functionality can be as simple as in this example. But as the complexity of the configuration grows,
automated tools to generate the configuration structure and basic consistency checks could be employed. These use
e.g. JSON Schema or its embedded form in OpenAPI 2 or 3.

A non-exhaustive list of these:

* https://github.com/xeipuuv/gojsonschema (JSON Schema)
* https://github.com/atombender/go-jsonschema (JSON Schema)
* https://github.com/ogen-go/ogen (OpenAPI 3.x)
* https://github.com/go-swagger/go-swagger (OpenAPI 2.0 / Swagger 2.0)

An example combining generation and templating can be found [here](examples/configSchema).


### Output & Secret Hiding

On program start, it is advisable to output the basic parameters controlling the following execution. However, many
configurations contain secrets, credentials for databases, access tokens etc. These should normally not be printed in
plain text to any location.

*templig* offers several possibilities to write the final configuration to a `Writer`:

   1. `To` writes the configuration completely, that is including secrets, to the given `Writer`.

      ```go
      c, _ := FromFile[Config]("my_config.yaml")
      c.To(os.Stdout)
      ```

      This program will produce the following, structurally identical output to the input configuration:

      ```yaml
      id:   23
      name: Interesting Name
      passes:
        - secretPass0
        - alternativePass1
      ```

   2. `ToSecretsHidden` writes the configuration, hiding secrets recognized using the `SecretRE` regular expression.
      The example of 1. will then become:

      ```go
      c, _ := FromFile[Config]("my_config.yaml")
      c.ToSecretsHidden(os.Stdout)
      ```

      With the new output to be:

      ```yaml
      id:   23
      name: Interesting Name
      pass: '*'
      ```

   3. `ToSecretsHiddenStructured` writes the configuration, hiding secrets, but letting their structure recognizable.
      The example of 1. will then become:

      ```go
      c, _ := FromFile[Config]("my_config.yaml")
      c.ToSecretsHiddenStructured(os.Stdout)
      ```

      With the new output to be:

      ```yaml
      id:   23
      name: Interesting Name
      pass:
        - '***********'
        - '****************'
      ```

Single secrets are always replaced by a string of `*` of equal length until a length of 32. Secrets longer than 32
characters are replaced by a string of `**` followed by the number of characters and a final `**`, e.g. `**42**`.
An example usage can be found [here](examples/templating/env).
