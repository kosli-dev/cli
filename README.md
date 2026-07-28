[![codecov](https://codecov.io/gh/kosli-dev/cli/branch/main/graph/badge.svg?token=Z4Y53XIOKJ)](https://codecov.io/gh/kosli-dev/cli)
[![Static Badge](https://img.shields.io/badge/provenance-blue?style=plastic&link=https%3A%2F%2Fapp.kosli.com%2Fkosli-public%2Fflows%2Fcli-release%2Ftrails%2F)](https://app.kosli.com/kosli-public/flows/cli-release/trails/)
[![Main](https://github.com/kosli-dev/cli/actions/workflows/main.yml/badge.svg)](https://github.com/kosli-dev/cli/actions/workflows/main.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

# Kosli CLI

The Kosli CLI records and queries software delivery events with [Kosli](https://www.kosli.com), giving you a tamper-evident record of how your software was built, tested, and deployed.

With it you can:

* **Fingerprint artifacts** — compute the SHA256 of files, directories, and OCI/Docker images.
* **Record attestations and evidence** — bind test results, security scans (Snyk, SonarQube), pull-request approvals, Jira issues, and custom evidence to your flows and trails.
* **Snapshot running environments** — report what is actually running in Kubernetes, ECS, Lambda, S3, Docker, Azure Web Apps, GCP Cloud Run, servers, and filesystem paths.
* **Query and assert compliance** — search and diff snapshots, and gate your CI/CD pipelines with `kosli assert` commands.

See the [documentation site](https://docs.kosli.com/) for the full command reference and usage guides.

## Installation

Install with whichever method suits your platform. After installing, run `kosli version` to verify.

### Install script (Linux / macOS)

```sh
curl -sSL https://raw.githubusercontent.com/kosli-dev/cli/main/install-cli.sh | sh
```

The script detects your OS and architecture and installs the matching release binary into a directory on your `PATH`. To install a specific version, pass the tag as an argument:

```sh
curl -sSL https://raw.githubusercontent.com/kosli-dev/cli/main/install-cli.sh | sh -s -- v2.35.0
```

### Homebrew (macOS / Linux)

```sh
brew install kosli-cli
```

Upgrade later with `brew upgrade kosli-cli`.

### npm

```sh
npm install -g @kosli/cli
```

> `npx @kosli/cli` is **not** supported — `npx` skips the optional platform dependency, so install the package first.

### Docker

```sh
docker run --rm ghcr.io/kosli-dev/cli:v2.35.0 version
```

Images are published to `ghcr.io/kosli-dev/cli` for each release tag.

### Build from source

```sh
make build      # produces ./kosli
./kosli version
```

See the [developer guide](/dev-guide.md) for full build details, including Windows.

`.deb` and `.rpm` packages are also published for each release. See the [install docs](https://docs.kosli.com/getting_started/install/) for the complete list of options.

## Documentation and links

* [Documentation site](https://docs.kosli.com/) for full details on usage.
* [Developer guide](/dev-guide.md) for details on working with the code in this repo.
* [Kosli main Trails](https://app.kosli.com/kosli-public/flows/cli/trails/)
* [Kosli release Trails](https://app.kosli.com/kosli-public/flows/cli-release/trails/)
