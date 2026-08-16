# helm compose up

Install releases and repositories defined in your `helm-compose.yaml`

## Usage

With no release names, this command installs all releases defined in your `helm-compose.yaml`, compares it to the latest applied revision, and uninstalls releases that have since been removed.

Pass one or more release names to install or upgrade only those releases and
their transitive `needs`. Other releases keep their previous applied state and
are not upgraded or uninstalled.

Releases in the same dependency group run concurrently. Use `--concurrency` to
limit the number of Helm processes. A failure cancels the other running
processes and prevents later dependency groups from starting.

```
helm compose up [RELEASE ...] [flags]

helm compose up wordpress2
helm compose up --release wordpress2
```

## Options

```
Flags:
      --concurrency int    Maximum concurrent Helm operations (0 means unlimited)
  -h, --help              help for up
  -r, --release strings   Release name to install or upgrade (can be specified multiple times)

Global Flags:
  -f, --file string   Compose configuration file
```
