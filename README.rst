.. image:: https://pkg.go.dev/badge/github.com/gofrontier-com/vertag.svg
    :target: https://pkg.go.dev/github.com/gofrontier-com/vertag
.. image:: https://github.com/gofrontier-com/vertag/actions/workflows/ci.yml/badge.svg
    :target: https://github.com/gofrontier-com/vertag/actions/workflows/ci.yml

=======
Vertag
=======

Vertag is a command line tool to manage versions of terraform modules with semver where the modules
are stored in the same repository.

.. contents:: Table of Contents
    :local:

-----
About
-----

Vertag has been built to assist with the management of terraform modules that are stored in the same
repository. It is designed to be used as part of a CI/CD pipeline.

--------
Download
--------

Binaries and packages of the latest stable release are available at `https://github.com/gofrontier-com/vertag/releases <https://github.com/gofrontier-com/vertag/releases>`_.

-----
Usage
-----

.. code:: bash

  $ vertag --help
  Vertag is the command line tool for managing terraform modules versioning

  Usage:
    vertag [flags]

  Flags:
    -e, --author-email string   Email of the commiter
    -n, --author-name string    Name of the commiter
    -d, --dry-run               Email of the commiter
    -h, --help                  Version
    -m, --modules-dir string    Directory of the modules
    -o, --output string         Output format (default "json")
    -u, --remote-url string     CI Remote URL
    -r, --repo string           Root directory of the repo
    -s, --short                 Print just the version number
    -v, --version               Version

------------
Contributing
------------

We welcome contributions to this repository. Please see `CONTRIBUTING.md <https://github.com/gofrontier-com/vertag/tree/main/CONTRIBUTING.md>`_ for more information.
