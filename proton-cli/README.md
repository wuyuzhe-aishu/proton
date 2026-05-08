# proton-cli

`proton-cli` is the Proton command-line tool for cluster deployment, configuration management, backup and recovery, Kubernetes maintenance, and offline package operations.

This README is written from the committed source code in this directory. It focuses on how to use the CLI, not on packaging artifacts that may exist outside the tracked code.

## 💬 Community

<div align="center">
<img src="./docs/qrcode.png" alt="KWeaver Community QR Code" width="30%"/>

Scan the QR code to join the KWeaver community and contact the group owner to obtain the Kweaver offline installation package
</div>

## Quick Start

### OS Compatibility
OS: OpenEuler 24 x86/arm, Kylin V10 SP3 x86/arm, RHEL8.10 X86, OpenEuler 23 x86/arm
Kernel: 4.11+
ARCH: amd64/arm64

### Offline Installation Quick Start

The following steps are for quickly installing Proton with a prebuilt `proton-cli` binary and an offline dependency package.

1. Download `proton-cli`

Download the latest release from GitHub Releases and choose the archive that matches your CPU architecture. The examples below use `v0.1.0`:

```bash
# amd64
wget https://github.com/kweaver-ai/proton/releases/download/v0.1.0/proton-cli-linux-amd64.tar.gz

# arm64
wget https://github.com/kweaver-ai/proton/releases/download/v0.1.0/proton-cli-linux-arm64.tar.gz

tar xf proton-cli-linux-<arch>.tar.gz
```

2. Download the Proton offline dependency package

Choose the offline package that matches your CPU architecture:

```bash
# amd64
proton-cli oras pull swr.cn-east-3.myhuaweicloud.com/kweaver-ai/offline-pkg/proton-offline-package:24653585065-amd64

# arm64
proton-cli oras pull swr.cn-east-3.myhuaweicloud.com/kweaver-ai/offline-pkg/proton-offline-package:24653585065-arm64
```

3. Install the offline package

```bash
tar xf proton-offline-package.tar
bash install.sh
```

4. Start the initialization service

Disable firewall

```bash
proton-cli server -ldebug -s proton-offline/service-package
```

Then open `http://<ip>:8888` in a browser to enter the initialization page.

5. Complete initialization

Set custom `MariaDB` and `Redis` passwords on the initialization page, then click `Complete` and wait for initialization to finish successfully.

6. Export an application offline package

Use the following command to download the charts and images defined in the specified `yaml` file and package them as an offline installation bundle:

```bash
proton-cli app export \
  -f release-manifest/x.y.z/kweaver-dip.yaml \
  --cache-dir /tmp/cache \
  --override-registry swr.cn-east-3.myhuaweicloud.com/kweaver-ai
```

7. Import an application offline package

Use the following command to import the charts and images from the offline application package into the built-in registry:

```bash
proton-cli app import \
  -i kweaver-dip-x.y.z-offline-package.tar \
  --auto --force
```

8. Install a specific product version

Use the following command to install the specified product version into the `kweaver` namespace:

```bash
proton-cli app install \
  -f release-manifest/x.y.z/kweaver-dip.yaml \
  -n kweaver
```

If you need to override values at install time, for example to disable optional capabilities:

```bash
proton-cli app install \
  -f release-manifest/x.y.z/kweaver-dip.yaml \
  -n kweaver \
  --set auth.enabled=false \
  --set businessDomain.enabled=false
```

`--set` overrides install-time values. See the later `app install` section for precedence rules and how it relates to manifest YAML.

### Requirements

- Linux environment
- Go 1.24 or newer to build from source
- A working Kubernetes client configuration at `~/.kube/config`
- For `apply`, a `service-package` directory that contains at least:
  - `charts/`
  - `images/`
- If you use `edit conf`, set `EDITOR`; otherwise `vi` is used

### Build

```bash
cd proton-cli
go build -o bin/proton-cli ./cmd/proton-cli
./bin/proton-cli --help
```

### Global Flags

Most commands accept these flags:

- `-l, --log-level`: log level, for example `info`, `debug`, `error`
- `-s, --service-package`: path to the `service-package` directory, default `service-package`
- `--service-package-eceph`: path to the `service-package-eceph` directory, default `service-package-eceph`
- `--cm-direct`: enable component-manage direct connection mode

## Common Workflows

### 1. Generate a cluster configuration template

Show one of the built-in templates:

```bash
proton-cli get template --type internal
proton-cli get template --type external
proton-cli get template --type perfrec
```

Create a file from a template:

```bash
proton-cli get template --type internal > cluster.yaml
```

Template types from the code:

- `internal`: deploy a local cluster
- `external`: deploy into an existing Kubernetes cluster
- `perfrec`: recommended configuration reference

### 2. Apply a cluster configuration

Apply a configuration file:

```bash
proton-cli apply -f cluster.yaml
```

Apply with an explicit deployment namespace:

```bash
proton-cli apply -f cluster.yaml -n proton
```

Apply with an explicit service package path:

```bash
proton-cli apply \
  -f cluster.yaml \
  -s /path/to/service-package \
  --service-package-eceph /path/to/service-package-eceph
```

Customize the HTTP/HTTPS ports for the `ingress-nginx` addon:

```yaml
cs:
  addons:
    - ingress-nginx
  addonsConfig:
    ingress-nginx:
      httpPort: 8080
      httpsPort: 8443
```

This maps to the `controller.containerPort.http/https` of the `ingress-nginx` chart.

Behavior confirmed from the code:

- `apply` reads the YAML file from `-f`
- `apply` loads package content from `service-package`
- if `-n` is given, it overrides the namespace from the config file
- the `deploy` field in the configuration file currently only supports `namespace` and `serviceaccount` fields; the old `mode` and `devicespec` fields have been removed and will be ignored if still present
- if a namespace is chosen, `apply` updates the local file `~/.proton-cli.yaml`
- after a successful apply, the cluster configuration is uploaded into Kubernetes

### 3. Show the current stored cluster configuration

Read the current configuration from Kubernetes:

```bash
proton-cli get conf
```

Read from a specific namespace:

```bash
proton-cli get conf -n proton
```

`get conf` reads the configuration from the Kubernetes Secret `proton-cli-config`.

### 4. Edit the stored configuration in Kubernetes

Open the stored configuration in your editor:

```bash
proton-cli edit conf
```

Edit a specific namespace:

```bash
proton-cli edit conf -n proton
```

Important:

- this command edits the Secret content directly
- it uses `$EDITOR`, or `vi` if `$EDITOR` is unset
- the command itself prints: `ONLY change secrets, not apply!`

### 5. Backup and recovery

Create a backup:

```bash
proton-cli backup create --resources all
```

Create a backup with an explicit name and retention:

```bash
proton-cli backup create \
  --backupname nightly-001 \
  --resources all \
  --ttl 3
```

List backups and inspect logs:

```bash
proton-cli backup list
proton-cli backup log
proton-cli backup directory
proton-cli backup schedule
```

Create a recovery from a backup:

```bash
proton-cli recover create \
  --from-backup nightly-001 \
  --resources all
```

List recoveries and inspect logs:

```bash
proton-cli recover list
proton-cli recover log
```

### 6. Build and install an offline package

Print the built-in manifest template:

```bash
proton-cli offline-package plan > manifest.yaml
```

Print a manifest template for a specific CPU architecture:

```bash
proton-cli offline-package plan --architecture arm64 > manifest.yaml
```

Build an offline package from a manifest:

```bash
proton-cli offline-package build --manifest manifest.yaml
```

The build command writes `proton-offline-package.tar` in the current directory.

Install an offline package:

```bash
proton-cli offline-package install proton-offline-package.tar
```

Keep the extracted working directory:

```bash
proton-cli offline-package install proton-offline-package.tar --remain
```

Behavior confirmed from the code:

- `build` reads `manifest.yaml` by default
- `plan` prints a generated manifest template to stdout and supports `--architecture` for `amd64` or `arm64`
- `install` extracts into `.proton-offline-package/` and runs `install.sh`

### 7. Export, import, install, and uninstall an application offline package

The `app` subcommand works with a VersionSet manifest to export, import, install, and uninstall an application package.

Export an application offline package:

```bash
proton-cli app export \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  --cache-dir /tmp/cache \
  --override-registry swr.cn-east-3.myhuaweicloud.com/kweaver-ai
```

If you want to set the output filename explicitly:

```bash
proton-cli app export \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -o kweaver-dip-0.5.0-offline-package.tar \
  --cache-dir /tmp/cache \
  --override-registry swr.cn-east-3.myhuaweicloud.com/kweaver-ai
```

When importing an application offline package, it is recommended to always add `--auto` so the built-in registry and ChartMuseum of the current Proton cluster are discovered automatically:

```bash
proton-cli app import -i kweaver-dip-0.5.0-offline-package.tar --auto
```

Overwrite existing charts if needed:

```bash
proton-cli app import -i kweaver-dip-0.5.0-offline-package.tar --auto --force
```

Install the application after import:

```bash
proton-cli app install \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -n kweaver
```

Only print the install plan without executing it:

```bash
proton-cli app install \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -n kweaver \
  --dry-run
```

Install with an explicit access address URL:

```bash
proton-cli app install \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -n kweaver \
  --access-address=https://1.1.1.1:8443/
```

Pass shared install values to all charts and use values to control optional dependencies:

```bash
proton-cli app install \
  -f release-manifests/0.6.0/kweaver-core.yaml \
  -n kweaver \
  --set auth.enabled=false \
  --set businessDomain.enabled=false
```

Uninstall the application:

```bash
proton-cli app uninstall \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -n kweaver
```

Common `app` flags:

- `app export`
- `-f, --manifest`: path to the VersionSet manifest
- `-o, --output`: output tar filename; if omitted, it is generated from the manifest
- `--cache-dir`: image export cache directory, useful when exporting the same image set repeatedly
- `--override-registry`: override `.image.registry` in chart values during export
- `--platform`: target export platform, auto-detected from the current host by default
- `--disable-dependencies`: only process the root manifest and skip recursive dependencies
- `--ignore-missing-images`: continue exporting even if some images cannot be pulled
- `app import`
- `--auto`: auto-discover the current Proton cluster registry and ChartMuseum, recommended by default
- `--force`: overwrite charts that already exist in ChartMuseum
- `--registry` / `--chartmuseum-url`: specify import targets manually; required if `--auto` is not used
- `app install`
- `-f, --file`: path to the VersionSet manifest used for installation
- `-n, --namespace`: target Kubernetes namespace, default `kweaver`
- `--timeout`: install timeout for each release, default `30m`
- `--dry-run`: only print the install plan
- `--config`: load values from a local config file when `proton-cli-config` cannot be read from the cluster
- `--image-registry`: override the image registry injected into values during install
- `--set`: pass install-time values as `key=value`; repeatable and supports dotted keys such as `auth.enabled=false`
- `--access-address`: override the access address using a full URL such as `https://1.1.1.1:8443/`; when the port is omitted, defaults are `443` for `https` and `80` for `http`
- `--access-host` / `--access-port` / `--access-scheme`: legacy split overrides; fields already present in values are preserved unless you explicitly pass these flags
- `app uninstall`
- `-f, --file`: path to the VersionSet manifest used for uninstall
- `-n, --namespace`: target Kubernetes namespace, default `kweaver`
- `--timeout`: uninstall timeout for each release, default `1m`
- `--dry-run`: only print the uninstall plan

Behavior confirmed from the code:

- `app export` reuses the `offline-package app export` implementation, including `--cache-dir` and `--override-registry`
- `app import` reuses the `offline-package app import` implementation, and `--auto` is recommended for regular use
- `app install` prints a completion message with the manifest and namespace after success
- `app install` expands manifest dependencies first, then computes install order using `stage: pre` and `dependsOn`
- `app install` merges cluster config values, `--set` overrides, and release-local `values` before calling Helm, with precedence `release.values > --set > cluster config`
- The relationship between `--set` and manifest YAML is:

```yaml
dependencies:
  - product: isf
    version: 0.6.0
    manifest: ./isf.yaml
    optional: true
    defaultEnabled: true
    enabledIf: auth.enabled

releases:
  mf-model-manager:
    chart: mf-model-manager
    version: 0.6.0
    values:
      auth:
        enabled: true
```

  Matching command:

```bash
proton-cli app install \
  -f release-manifests/0.6.0/kweaver-core.yaml \
  --set auth.enabled=false \
  --set businessDomain.enabled=false
```

  Semantics:
  `--set auth.enabled=false` is injected as a global values override for every chart; if a release defines `auth.enabled` again under manifest `values`, that release-local value wins.
  `dependencies[].enabledIf: auth.enabled` reads from the same merged install-time values; when the value is `false`, the dependency is skipped, and when the value is absent, the planner falls back to `defaultEnabled`.
- `dependsOn` means a ready dependency between releases in the same manifest; the dependency release is installed and waited for first
- optional dependencies can be controlled by manifest `enabledIf` plus `--set`, for example:

```yaml
dependencies:
  - product: isf
    version: 0.6.0
    manifest: ./isf.yaml
    defaultEnabled: true
    enabledIf: auth.enabled
```
- `app uninstall` runs in reverse install order

### 8. Kubernetes utilities

Show the current Kubernetes-related state:

```bash
proton-cli kubernetes show
```

Upgrade Calico:

```bash
proton-cli kubernetes calico upgrade <version>
```

The command validates the target version against a built-in supported version list before starting the upgrade.

### 9. Shell completion and version

Print version info:

```bash
proton-cli version
```

Enable shell completion:

```bash
proton-cli completion bash
proton-cli completion zsh
proton-cli completion fish
proton-cli completion powershell
```

## Command Overview

Root commands exposed by the committed code include:

- `apply`: apply a cluster configuration file
- `get`: show current configuration or built-in templates
- `edit`: edit stored configuration in Kubernetes
- `backup`: create and inspect backups
- `recover`: create and inspect recoveries
- `offline-package`: print a manifest template, build packages, install packages
- `app`: export, import, install, and uninstall VersionSet-based application packages
- `kubernetes`: show Kubernetes state and manage Calico
- `completion`: generate shell completion
- `version`: print version information
- `precheck`: check node environment before install
- `reset`: reset a Proton cluster
- `migrate`: migrate components deployed by other programs
- `check`: check Proton runtime health
- `images`: manage images
- `push-images`: push images to a repository
- `push-charts`: push charts to a repository
- `package`: packaging-related commands
- `component`: data component management
- `delete-images`: delete images to reclaim disk space
- `server`: run the CLI server mode
- `alpha`: experimental commands

Use help on demand:

```bash
proton-cli --help
proton-cli <command> --help
```

## Runtime State and Default Locations

The committed code uses these default locations and names:

- Kubernetes client config: `~/.kube/config`
- local Proton CLI environment file: `~/.proton-cli.yaml`
- default Proton CLI config namespace: `proton`
- default Proton resource namespace: `resource`
- stored cluster configuration Secret: `proton-cli-config`
- cluster configuration key inside the Secret: `ClusterConfiguration`

In practice:

- `get conf` and `edit conf` read from `proton-cli-config`
- `apply` may update `~/.proton-cli.yaml`
- namespace-sensitive commands usually default to the namespace recorded in `~/.proton-cli.yaml`, or `proton` if the file does not exist

## Notes

- This README intentionally does not assume the repository already contains prebuilt `service-package` assets.
- If you are unsure about a command, prefer the built-in help output over older scripts or placeholders.
