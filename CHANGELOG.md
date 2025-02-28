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

- use parse structure of YAML documents for secrets hiding
- documentation fixes
- node based function `HideSecrets`, to hide secret in configurations

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