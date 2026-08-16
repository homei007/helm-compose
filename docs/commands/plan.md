# helm compose plan

Preview the release operations Helm Compose would perform without changing the
cluster. `helm compose diff` is an alias for this command.

## Usage

```
helm compose plan [RELEASE ...] [flags]

helm compose plan
helm compose diff --release app
```

The output is grouped in execution order. Releases in the same group can run
concurrently. Actions are:

| Action    | Meaning |
| --------- | ------- |
| install   | The release is not present in the latest applied Compose revision. |
| upgrade   | The release configuration differs from the latest applied revision. |
| sync      | The Compose configuration is unchanged, but `up` will still ask Helm to reconcile the release. |
| uninstall | The release exists in the latest applied revision but was removed from the current file. |

This is a Compose-state comparison. It does not render charts or query the
cluster for live-manifest drift.

When release names are supplied, their transitive `needs` are included and
removed releases are not planned for uninstall.

## Options

```
Flags:
  -h, --help              help for plan
  -r, --release strings   Release name to plan (can be specified multiple times)

Global Flags:
  -f, --file string   Compose configuration file
```
