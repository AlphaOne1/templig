<!-- SPDX-FileCopyrightText: 2026 The templig contributors.
     SPDX-License-Identifier: MPL-2.0
-->

Security Assurance Case
=======================

Threat Model & Trust Boundaries
-------------------------------

*templig* defines clear boundaries for data trust:

* __Trust Boundary__:

  The transition from the Input Layer (files, environment variables) to the
  *templig* Core.

* __Threats__:

  We assume all external inputs (templates and environment variables) are
  potentially untrusted or malformed.

* __Mitigation__:

  Data is never executed as code. It is processed by the Go text/template engine
  (for data templating, not HTML sanitization) and subsequently strictly
  validated by the Config Validator against a defined schema.

Secure Design Principles (Saltzer & Schroeder)
----------------------------------------------

We applied the following principles:

* __Economy of Mechanism__:

  The core is kept small and uses Go’s standard library where possible to
  minimize the attack surface. Libraries such as
  [Sprig](https://github.com/Masterminds/sprig) have proven themselves in
  practice e.g., [helm](https://github.com/helm/helm).

* __Least Privilege__:

  *templig* requires no network access, no root privileges, and operates
  entirely in-memory.

* __Fail-safe Defaults__:

  If a template cannot be parsed or a validation fails, *templig* returns an
  error, preventing the use of "half-baked" or insecure configurations.

Countering Common Weaknesses (CWE / OWASP)
------------------------------------------

*templig* is designed to mitigate common implementation weaknesses:

* __Injection (CWE-74)__:

  By using a structured validation step after rendering, it is ensured that
  template injections cannot result in invalid or malicious configuration
  structures. Further is the input into *templig* considered to stem from an
  authorized person. It is _not_ intended to be used for untrusted end-user
  inputs.

* __Information Exposure (CWE-200)__:

  The Secret Hiding layer ensures that sensitive data (like passwords from
  environment variables) is masked in logs and standard output. It is
  obligation of the programmer to use the secured functions. Non-hiding
  functions are intentionally provided for debugging purposes.

* __Improper Input Validation (CWE-20)__:

  Every rendered configuration is parsed and validated against the expected
  target structure (e.g., YAML/JSON schema) before being returned to the
  program. A programmer may further choose to add extended validation
  functionality.

Evidence of Assurance
---------------------

* __Static Analysis__:

  Continuous scanning via [golangci-lint](https://golangci-lint.run)
  (including gosec).

* __Test Coverage__:

  High statement coverage is tracked in CI to increase confidence that
  error-handling paths (fail-safes) are exercised.

* __Build Integrity__:

  SLSA Level 3 provenance provides build traceability and stronger assurance
  that distributed artifacts were produced by the documented build process from
  the declared source and inputs.