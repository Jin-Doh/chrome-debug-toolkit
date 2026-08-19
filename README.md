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

## Development checks

```bash
go fmt ./...
go test ./...
go vet ./...
go build ./...
```
