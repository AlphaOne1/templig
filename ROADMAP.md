<!-- SPDX-FileCopyrightText: 2026 The templig contributors.
     SPDX-License-Identifier: MPL-2.0
-->

Roadmap
=======

Although *templig* is considered production-ready, there are some points that
could be improved upon:

Upcoming
--------

* __Soon: Overlay Replacement on Arrays__

  Arrays are currently merged, following the same argument as merging associative
  arrays or objects. However, replacing the content of an array could be desirable
  and could be implemented using a custom YAML annotation, e.g., `!replace`.

  This would also allow for replacing an existing base object with `null`, as this
  is currently not possible due to a type mismatch.

* __In Design: REST Calls__

  REST calls are a common way to fetch dynamic configuration in cloud-native
  environments. Integrating REST support will allow templig to pull data from
  service registries or metadata APIs (like AWS/GCP Metadata). This feature is
  still in design and requires further research and implementation.

* __Long shot: Database Access__

  In container environments there are often databases or at least central
  datastores available that contain configuration data. Examples are:

  * [etcd](https://etcd.io)
  * [Hashicorp Vault](https://www.hashicorp.com/de/products/vault)
  * relational databases
    ([PostgreSQL](https://www.postgresql.org),
     [MariaDB](https://mariadb.org),
     [SQLite](https://sqlite.org), ...)
  * other key-value stores ([Redis](https://redis.io), [Memcached](https://memcached.org),...)

  If we could also get information from these, that also would maximize the
  versatility of *templig*. As database drivers might themselves import a huge
  amount of dependencies, it should be made an optional feature. It is to be
  defined, if this would be for the programmer to decide, or if it is possible
  to manage it via plugins at runtime.

Community & Documentation
-------------------------

* __Improved Documentation__

  The documentation of *templig* is surely not perfect. If you are a new user
  and find a question not properly answered, feel free to ask, or even better,
  send improvements.


* __Extended Examples__

  The examples try to show each aspect of *templig*. Specific use-cases may be
  more complex than the single-feature-centered original examples. If you
  encounter an interesting use-case that can be simplified enough to serve as an
  example, you are welcome to contribute it.
