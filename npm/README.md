# NPM Packaging

This directory contains the npm package structure for distributing the Kosli CLI via npm, following the same pattern used by [esbuild](https://github.com/evanw/esbuild).

## Structure

```
npm/
├── wrapper/              # @jbrejner/cli — the package users install
│   ├── package.json      # declares optionalDependencies for all platforms
│   ├── bin/kosli         # JS shim that detects the platform and runs the binary
│   └── install.js        # postinstall script that validates the binary
├── cli-linux-x64/        # @jbrejner/cli-linux-x64
│   ├── package.json      # declares os/cpu fields for platform filtering
│   └── bin/kosli         # the native binary — see below
├── cli-linux-arm64/      # @jbrejner/cli-linux-arm64
│   ├── package.json      # declares os/cpu fields for platform filtering
│   └── bin/kosli         # the native binary — see below
├── cli-linux-arm/        # @jbrejner/cli-linux-arm
│   ├── package.json      # declares os/cpu fields for platform filtering
│   └── bin/kosli         # the native binary — see below
├── cli-darwin-x64/       # @jbrejner/cli-darwin-x64
│   ├── package.json      # declares os/cpu fields for platform filtering
│   └── bin/kosli         # the native binary — see below
├── cli-darwin-arm64/     # @jbrejner/cli-darwin-arm64
│   ├── package.json      # declares os/cpu fields for platform filtering
│   └── bin/kosli         # the native binary — see below
└── cli-win32-x64/        # @jbrejner/cli-win32-x64
│   ├── package.json      # declares os/cpu fields for platform filtering
│   └── bin/kosli[.exe]   # the native binary — see below
└── cli-win32-arm64/      # @jbrejner/cli-win32-arm64
    ├── package.json      # declares os/cpu fields for platform filtering
    └── bin/kosli[.exe]   # the native binary — see below
```

## How it works

Users install a single package:

```sh
npm install @jbrejner/cli
```

or if using in continuous integration you can install globally:

```sh
npm install -g @jbrejner/cli
```

npm resolves the `optionalDependencies` declared in the wrapper's `package.json` and installs only the platform-specific package that matches the current OS and CPU architecture — all non-matching packages are silently skipped. The wrapper's `bin/kosli` JS shim then locates the binary inside the installed platform package and executes it.

## The `bin/` directories are populated by goreleaser

The platform package `bin/` directories are **not committed to git**. They are populated automatically during the release process by a post-build hook in [`.goreleaser.yml`](../.goreleaser.yml):

```yaml
hooks:
  post:
    - cmd: >-
        bash -c '
        OS="{{ .Os }}";
        ARCH="{{ .Arch }}";
        [ "$OS" = "windows" ] && OS="win32";
        [ "$ARCH" = "amd64" ] && ARCH="x64";
        EXT="";
        [ "{{ .Os }}" = "windows" ] && EXT=".exe";
        mkdir -p npm/cli-${OS}-${ARCH}/bin &&
        cp "{{ .Path }}" npm/cli-${OS}-${ARCH}/bin/kosli${EXT} &&
        chmod +x npm/cli-${OS}-${ARCH}/bin/kosli${EXT}'
```

This hook runs once per build target immediately after goreleaser compiles the binary. It applies the following naming conventions:

| goreleaser | npm package dir |
|------------|-----------------|
| `linux`    | `linux`         |
| `darwin`   | `darwin`        |
| `windows`  | `win32`         |
| `amd64`    | `x64`           |
| `arm64`    | `arm64`         |
| `arm`      | `arm`           |

Windows binaries are copied as `kosli.exe`; all others as `kosli`. The `windows/arm` combination is excluded from builds.

The `before` hooks in `.goreleaser.yml` clean up stale artifacts before each build run:

```yaml
before:
  hooks:
    - rm -rf npm/cli-*/bin
    - find npm -name "*.tgz" -delete
```

## Publishing

Packages are published to the [npm public registry](https://registry.npmjs.org). Platform packages must be published before the wrapper, since the wrapper's `optionalDependencies` references them by version. After a goreleaser build has populated the `bin/` directories:

```sh
# Publish platform packages first
(cd npm/cli-linux-x64    && npm publish)
(cd npm/cli-linux-arm64  && npm publish)
(cd npm/cli-linux-arm    && npm publish)
(cd npm/cli-darwin-x64   && npm publish)
(cd npm/cli-darwin-arm64 && npm publish)
(cd npm/cli-win32-x64    && npm publish)
(cd npm/cli-win32-arm64  && npm publish)

# Then publish the wrapper
(cd npm/wrapper && npm publish)
```

Each package directory contains an `.npmrc` that sets the auth token:

```
//registry.npmjs.org/:_authToken=${MY_LOCAL_NPM_TOKEN}
```

## Automated Publishing with npm-publish.sh

The `scripts/npm-publish.sh` script automates the npm packaging and publishing process. It injects the version into all `package.json` files, packs each package into a `.tgz`, and optionally publishes them.

### Usage

```bash
scripts/npm-publish.sh <version> [--dry-run]
```

### Arguments

- `<version>`: Required. A SemVer string — either `X.Y.Z` (stable) or `X.Y.Z-TAG` (pre-release).
- `--dry-run` (optional second argument): Pack packages but skip publishing.

### Behavior

1. Injects `<version>` into the `version` field of all `package.json` files.
2. Updates the `optionalDependencies` version references in `npm/wrapper/package.json` to match.
3. Runs `npm pack` on each platform package, then on the wrapper.
4. Unless `--dry-run` is set, runs `npm publish --tag <tag>` on each package.

The dist-tag is determined by the version format:

| Version format | npm dist-tag |
|----------------|--------------|
| `X.Y.Z`        | `latest`     |
| `X.Y.Z-*`      | `snapshot`   |

### Integration with GoReleaser

GoReleaser calls this script automatically via the `after` hook once all platform binaries have been built and copied into the `bin/` directories:

```yaml
after:
  hooks:
    - cmd: bash scripts/npm-publish.sh "{{ .Version }}" ...
      output: true
```

The script output is surfaced in the goreleaser log (`output: true`).

## Versioning

All packages share the same version number. When releasing, `npm-publish.sh` updates it automatically in all eight `package.json` files — the seven platform packages and the wrapper — as well as the `optionalDependencies` version pins in `npm/wrapper/package.json`. There is no need to edit these files manually.
