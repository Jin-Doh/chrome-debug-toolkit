#!/usr/bin/env bash
set -euo pipefail

workflow_dir="${1:-.github/workflows}"

if [[ ! -d "$workflow_dir" ]]; then
    echo "workflow directory not found: $workflow_dir" >&2
    exit 1
fi

shopt -s nullglob
workflow_files=("$workflow_dir"/*.yml "$workflow_dir"/*.yaml)
shopt -u nullglob

if [[ "${#workflow_files[@]}" -eq 0 ]]; then
    echo "no workflow files found in $workflow_dir" >&2
    exit 1
fi

# Git refs are repository-scoped. Treat all actions from the same owner/repo as
# one version family so sub-actions such as CodeQL init/autobuild/analyze cannot
# drift onto different commits. Also preserve the repository's full-SHA pinning
# policy for third-party actions.
awk '
function fail(message) {
    print message > "/dev/stderr"
    status = 1
}

{
    line = $0
    if (line !~ /^[[:space:]]*(-[[:space:]]*)?uses:[[:space:]]*/) {
        next
    }

    sub(/^[[:space:]]*(-[[:space:]]*)?uses:[[:space:]]*/, "", line)
    sub(/[[:space:]#].*$/, "", line)

    first = substr(line, 1, 1)
    last = substr(line, length(line), 1)
    if ((first == "\"" && last == "\"") || (first == "\047" && last == "\047")) {
        line = substr(line, 2, length(line) - 2)
    }

    at = index(line, "@")
    if (at == 0) {
        next
    }

    action = substr(line, 1, at - 1)
    ref = substr(line, at + 1)

    count = split(action, parts, "/")
    if (count < 2 || parts[1] == "." || parts[1] == "docker:") {
        next
    }

    repository = parts[1] "/" parts[2]
    location = FILENAME ":" FNR
    found = 1

    if (length(ref) != 40 || ref !~ /^[0-9a-fA-F]+$/) {
        fail(location ": " repository " is not pinned to a full commit SHA: " ref)
    }

    if ((repository in refs) && refs[repository] != ref) {
        fail(location ": inconsistent ref for " repository)
        print "  first: " refs[repository] " at " locations[repository] > "/dev/stderr"
        print "  found: " ref " at " location > "/dev/stderr"
    } else if (!(repository in refs)) {
        refs[repository] = ref
        locations[repository] = location
    }
}

END {
    if (!found) {
        print "no external GitHub Action references found" > "/dev/stderr"
        exit 1
    }

    if (status != 0) {
        exit status
    }

    print "GitHub Action references are SHA-pinned and synchronized by source repository."
}
' "${workflow_files[@]}"
