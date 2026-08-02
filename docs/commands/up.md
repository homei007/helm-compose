# helm compose up

Install releases and repositories defined in your `helm-compose.yaml`

## Usage

With no release names, this command installs all releases defined in your `helm-compose.yaml`, compares it to the latest applied revision, and uninstalls releases that have since been removed.

Pass one or more release names to install or upgrade only those releases. Unselected releases keep their previous applied state and are not upgraded or uninstalled.

```
helm compose up [RELEASE ...] [flags]

helm compose up wordpress2
helm compose up --release wordpress2
```

## Options

```
Flags:
  -h, --help              help for up
  -r, --release strings   Release name to install or upgrade (can be specified multiple times)

Global Flags:
  -f, --file string   Compose configuration file
```
