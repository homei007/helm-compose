# Helm Compose (Community-Maintained Fork)

This repository is a community-maintained continuation of the original
[seacrew/helm-compose](https://github.com/seacrew/helm-compose), which is no
longer actively maintained. It preserves the `helm compose` command and the
existing Helm Compose file format while accepting fixes and improvements.

This project is not affiliated with or endorsed by Seacrew or the Helm project.

Current development focuses on release selection, concurrent processing, and
keeping the plugin reliable and easy to install.

![helm-compose-banner](https://user-images.githubusercontent.com/18513179/240495789-e76890d3-f0f9-48b9-9d18-89e53effe65b.png)

[![Build Status](https://github.com/homei007/helm-compose/actions/workflows/build.yaml/badge.svg)](https://github.com/homei007/helm-compose/actions/workflows/build.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/homei007/helm-compose)](https://goreportcard.com/report/github.com/homei007/helm-compose)

Helm Compose is a tool for managing multiple releases of one or many different Helm charts. It is heavily inspired by Docker Compose and is an extension of the package manager idea behind Helm itself. It allows for full configuration-as-code capabilities in an single yaml file.

## Installation

It is requirement to use helm v3.10.0+.

The fork is currently in active development and does not have a published
binary release yet. For local development from this repository:

```
make install
```

After a fork release is published, install it with:

```
helm plugin install https://github.com/homei007/helm-compose
```

## Quick Start Guide

Helm Compose makes it easy to define a list of Releases and all necessary Repositories for the charts you use in a single compose file.

Install your releases:

```bash
$ helm compose up -f helm-compose.yaml
```

Install or upgrade only selected releases:

```bash
$ helm compose up -f helm-compose.yaml wordpress wordpress2
```

Uninstall your releases

```bash
$ helm compose down -f helm-compose.yaml
```

Uninstall only selected releases:

```bash
$ helm compose down -f helm-compose.yaml wordpress2
```

A Helm Compose file looks something like this:

```yaml
apiVersion: 1.1

storage:
  name: mycompose
  type: local # default
  path: .hcstate # default

releases:
  wordpress:
    chart: bitnami/wordpress
    chartVersion: 14.3.2
  wordpress2:
    chart: bitnami/wordpress
    chartVersion: 15.2.22
    namespace: homepage
    createNamespace: true
  postgres:
    chart: bitnami/postgresql
    chartVersion: 12.1.9
    namespace: database
    createNamespace: true

repositories:
  bitnami: https://charts.bitnami.com/bitnami
```

Check out the [examples](https://github.com/homei007/helm-compose/tree/main/examples) directory.

## Documentation

Check out the complete [documentation](https://homei007.github.io/helm-compose/).

## License and attribution

Helm Compose is distributed under the Apache License 2.0. This fork retains
the original license, copyright notices, and third-party license files. Please
see [LICENSE](LICENSE) and [LICENSES](LICENSES/) for details.
