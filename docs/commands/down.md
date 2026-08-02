# helm compose down

Uninstall releases defined in your `helm-compose.yaml`

## Usage

With no release names, this command uninstalls all releases from the previous applied revision if one exists. Otherwise, it uninstalls the releases defined in your current `helm-compose.yaml`.

Pass one or more release names to uninstall only those releases. The applied revision is updated so the remaining releases are left intact.

```
helm compose down [RELEASE ...] [flags]

helm compose down wordpress2
helm compose down --release wordpress2
```

## Options

```
Flags:
  -h, --help              help for down
  -r, --release strings   Release name to uninstall (can be specified multiple times)

Global Flags:
  -f, --file string   Compose configuration file
```
