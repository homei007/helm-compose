# Helm Compose (Community-Maintained Fork)

This repository is a community-maintained continuation of the original
[seacrew/helm-compose](https://github.com/seacrew/helm-compose), which is no
longer actively maintained. It preserves the `helm compose` command and the
existing Helm Compose file format while accepting fixes and improvements.

This project is not affiliated with or endorsed by Seacrew or the Helm project.

Current development focuses on reliable multi-release operations, concurrent
processing, Helm 3 and Helm 4 compatibility, and making the plugin easy to
install and maintain.

![helm-compose-banner](https://user-images.githubusercontent.com/18513179/240495789-e76890d3-f0f9-48b9-9d18-89e53effe65b.png)

[![Build Status](https://github.com/homei007/helm-compose/actions/workflows/build.yaml/badge.svg)](https://github.com/homei007/helm-compose/actions/workflows/build.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/homei007/helm-compose)](https://goreportcard.com/report/github.com/homei007/helm-compose)

Helm Compose is a tool for managing multiple releases of one or many different Helm charts. It is heavily inspired by Docker Compose and is an extension of the package manager idea behind Helm itself. It allows for full configuration-as-code capabilities in an single yaml file.

## Installation

Helm Compose currently supports Helm v3.10.0 and newer. Helm v4 support is
being validated; normal CLI-plugin workflows are expected to work, while
post-renderer configurations still need a Helm v4-specific compatibility
update.

Install the latest published release with:

```
helm plugin install https://github.com/homei007/helm-compose --version 1.4.0
```

For local development from this repository:

```
make install
```

After a fork release is published, install it with:

```
helm plugin install https://github.com/homei007/helm-compose
```

## Roadmap

This roadmap describes the next priorities for the community-maintained fork.
Actual timing depends on testing, user feedback, and contributor capacity.

### Compatibility and reliability

- Publish a Helm 3/Helm 4 compatibility matrix and run both versions in CI.
- Migrate the plugin manifest to Helm 4's versioned subprocess schema.
- Update post-renderer handling for Helm 4 while preserving Helm 3 behavior.
- Improve failure reporting and cancellation when concurrent releases are
  processed.
- Expand integration coverage for local and S3 state storage, namespaces,
  dependencies, and repeatable upgrades.

### User experience

- Add a dry-run or plan mode that shows the actions Helm Compose will take.
- Improve selective release operations, dependency ordering, and progress
  output for larger compose files.
- Document migration paths, common failure modes, and tested Helm/Kubernetes
  combinations.
- Add more practical examples for production-style repositories and values
  management.

### Sustainable maintenance

- Keep automated tests, release packaging, checksums, and documentation
  publishing on every release.
- Add issue and pull-request templates to make support and contribution easier.
- Collect real-world feedback from Helm users and prioritize changes that
  preserve the existing compose-file format.
- Make security and dependency updates part of the normal maintenance cycle.

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
