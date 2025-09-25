<!-- SPDX-FileCopyrightText: 2025 The templig contributors.
     SPDX-License-Identifier: MPL-2.0
-->

Release 0.8.3
=============

- changed unmaintained gopkg.in/yaml.v3 to go.yaml.in/yaml/v4
- dependency updates

Release 0.8.2
=============

- documentation clarification for `Config.Get`
- dependency updates
- redacted code using AI to point out potential

Release 0.8.1
=============

- configured linter
- adhered to linter comments

Release 0.8.0
=============

- add an example for config object generation from JSON schema
- more documentation in examples
- updated dependencies
- add golangci-lint checks

Release 0.7.0
=============

- give user the possibility to disallow additional templig template functions or add their own ones

Release 0.6.2
=============

- update x/crypto dependency 0.34.0 -> 0.35.0

Release 0.6.1
=============

- fix not printing using secret hiding
- documentation for configuration printing and secret hiding
- extended templating/env example to print configuration hiding secrets
- improve secret detection mechanism

Release 0.6.0
=============

- deprecated `FromFiles`, replaced by `FromFile`
- multi-reader support for `From`
- multi-file support for `FromFile`

Release 0.5.0
=============

- added `arg` and `hasArg` template functions
- documentation structure improved
- restructured examples to specifically address templating functions

Release 0.4.1
=============

- added documentation for `FromFiles`

Release 0.4.0
=============

- use parse structure of YAML documents for secret hiding
- documentation fixes
- node-based function `HideSecrets`, to hide secret in configurations

Release 0.3.0
=============

- update dependencies
- added node-based YAML merge function `MergeYAMLNodes` to prevent 
  deserialization inaccuracies
- added `FromFiles` to facilitate reading from multiple input files
  that overlay each other in order

Release 0.2.0
=============

- added `read` function to read file contents
- added `Validator` interface so that configurations can
  be checked for validity when they are being loaded

Release 0.1.0
=============

Initial release

- added base config reader
- added support to use text/template
- added sprig functions for templates