Roadmap
=======

Although *templig* is considered production ready, there are some points that
could be improved upon:

Upcoming
--------

* __Soon: Command Line Argument Support__

  Currently *templig* does not include dedicated support for command line
  arguments. It is the last missing piece to cover classical possible input
  vectors. If we could provide command line argument read access to templates in
  an easy-to-use way, it would further improve *templig*'s applicability.

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
  * other key-value stores ([Redis](https://redis.io), [Memcached](https://memcached.org)),...)

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
