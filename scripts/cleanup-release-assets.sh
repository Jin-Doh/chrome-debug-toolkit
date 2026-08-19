#!/usr/bin/env bash
set -euo pipefail

repo="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
keep="${RELEASE_RETENTION_COUNT:-5}"

if [[ ! "$keep" =~ ^[1-9][0-9]*$ ]]; then
    echo "RELEASE_RETENTION_COUNT must be a positive integer: $keep" >&2
    exit 2
fi

stale_release_ids="$(
    gh api --paginate --slurp "repos/$repo/releases?per_page=100" |
        jq -r --argjson keep "$keep" '
            [.[][] | select(.draft == false)]
            | sort_by(.published_at // .created_at)
            | reverse
            | .[$keep:]
            | .[].id
        '
)"

while IFS= read -r release_id; do
    [[ -z "$release_id" ]] && continue
    asset_ids="$(
        gh api --paginate --slurp "repos/$repo/releases/$release_id/assets?per_page=100" |
            jq -r '[.[][]] | .[].id'
    )"
    while IFS= read -r asset_id; do
        [[ -z "$asset_id" ]] && continue
        gh api --method DELETE "repos/$repo/releases/assets/$asset_id" >/dev/null
        echo "Deleted stale release asset $asset_id from release $release_id"
    done <<EOF
$asset_ids
EOF
done <<EOF
$stale_release_ids
EOF
