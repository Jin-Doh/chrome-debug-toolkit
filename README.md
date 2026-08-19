# Chrome Debug Toolkit

`cdt` captures Chrome NetLog evidence in an isolated Chrome profile. It does
not attach flags to, terminate, or replace the user's normal Chrome process.

## Requirements

- Go 1.26 or newer
- Google Chrome or Chromium
- macOS is the primary supported platform; common Linux installations are
  detected as well

## Build and install

```bash
go install github.com/jin-doh/chrome-debug-toolkit/cmd/cdt@latest
# or from this checkout
go build -o cdt ./cmd/cdt
```

## Install the latest release

Release tags use semantic versioning with a `v` prefix. Pushing a tag such as
`v0.1.0` runs the release pipeline, publishes archives and checksums, and
updates the GitHub `latest` release:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

Install the matching prebuilt binary into `$HOME/.local/bin/`:

```bash
set -euo pipefail

case "$(uname -s):$(uname -m)" in
  Darwin:arm64)  asset="cdt-darwin-arm64.tar.gz" ;;
  Darwin:x86_64) asset="cdt-darwin-amd64.tar.gz" ;;
  Linux:aarch64) asset="cdt-linux-arm64.tar.gz" ;;
  Linux:x86_64)  asset="cdt-linux-amd64.tar.gz" ;;
  *) echo "Unsupported platform: $(uname -s):$(uname -m)" >&2; exit 1 ;;
esac

base="https://github.com/Jin-Doh/chrome-debug-toolkit/releases/latest/download"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

curl --fail --location --silent --show-error \
  "$base/$asset" -o "$tmp/$asset"
tar -xzf "$tmp/$asset" -C "$tmp"
mkdir -p "$HOME/.local/bin"
install -m 0755 "$tmp/cdt" "$HOME/.local/bin/cdt"
```

Ensure the install directory is on `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
cdt version
```

Each release includes archives for macOS and Linux plus `checksums.txt`.
The release pipeline retains the assets of the five most recent published
releases and removes assets from older releases. Release records, tags, and
release notes are preserved.

## Commands

Start a capture browser. The command returns after Chrome starts, so leave that
window open while reproducing the problem:

```bash
cdt netlog
cdt netlog https://example.com
```

The normal Chrome instance remains untouched. The managed profile, session
metadata, stdout/stderr, and NetLog paths are recorded separately.

```bash
cdt ps                 # list Chrome processes and mark managed ones
cdt sessions           # list sessions; update stale running sessions to exited
cdt inspect latest     # inspect the newest NetLog without loading the file whole
cdt inspect SESSION_ID
cdt inspect /path/to/netlog.json
cdt kill               # SIGTERM only managed Chrome processes
cdt kill --force       # SIGKILL only managed Chrome processes
cdt clean 7            # remove sessions and their NetLogs older than 7 days
cdt doctor             # inspect local Chrome, storage, ps, and CDP availability
```

`cdt kill` matches the toolkit's managed `--user-data-dir` argument. It never
uses a broad process name kill, so the normal Chrome profile is not targeted.

## Storage

By default, macOS paths are:

```text
~/Library/Application Support/chrome-debug-toolkit/
├── profiles/netlog/
└── sessions/<SESSION_ID>/
    ├── session.json
    ├── chrome.stdout.log
    └── chrome.stderr.log

~/Downloads/chrome-netlogs/chrome-netlog-<SESSION_ID>.json
```

Override paths for a local installation or isolated tests:

```bash
export CHROMEPROBE_CHROME="/path/to/Google Chrome"
export CHROMEPROBE_DATA_DIR="$PWD/.cdt-data"
export CHROMEPROBE_NETLOG_DIR="$PWD/.cdt-netlogs"
```

The persistent managed profile can contain cookies and other browsing state.
NetLog can contain URLs, headers, endpoints, and other sensitive network
metadata. Review captures before sharing them.

## Compatibility wrappers

The repository keeps the existing habits as thin wrappers:

```bash
./chromenetlog https://example.com
./chromenetlog-clean 7
```

They delegate to `cdt`; they do not contain a second Chrome implementation.

## Development quality gate

The repository's local quality gate is:

```bash
go fmt ./...
golangci-lint run ./...
go vet ./...
go test ./...
go build ./...
./scripts/check-coverage.sh
```

`check-coverage.sh` runs the full test suite with an atomic coverage profile and
fails below the project-wide 85% statement threshold. Set
`COVERAGE_THRESHOLD` or `COVERAGE_PROFILE` to override the defaults for local
experiments; CI uses the standard 85% threshold.

`gopls` is an editor/LSP tool and is intentionally not a runtime dependency.
