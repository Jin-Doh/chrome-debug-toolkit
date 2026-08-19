#!/usr/bin/env bash
set -euo pipefail

threshold="${COVERAGE_THRESHOLD:-85.0}"
profile="${COVERAGE_PROFILE:-coverage.out}"

go test ./... -covermode=atomic -coverprofile="$profile"
actual="$(go tool cover -func="$profile" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')"

if [[ -z "$actual" ]]; then
    echo "coverage total was not reported" >&2
    exit 1
fi

awk -v actual="$actual" -v threshold="$threshold" 'BEGIN {
    if (actual + 0 < threshold + 0) {
        printf "coverage %.1f%% is below required %.1f%%\n", actual, threshold > "/dev/stderr"
        exit 1
    }
    printf "coverage %.1f%% meets required %.1f%%\n", actual, threshold
}'
