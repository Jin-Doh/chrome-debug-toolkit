#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:?VERSION is required (without the leading v)}"
dist="${DIST_DIR:-dist}"

rm -rf "$dist"
mkdir -p "$dist"
dist_abs="$(cd "$dist" && pwd)"

for target in darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/amd64; do
    os="${target%/*}"
    arch="${target#*/}"
    asset="cdt-${os}-${arch}"
    staging="$(mktemp -d)"
    trap 'rm -rf "$staging"' EXIT

    binary="$staging/cdt"
    if [[ "$os" == "windows" ]]; then
        binary+=".exe"
    fi

    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build \
        -trimpath \
        -ldflags="-s -w -X main.version=$version" \
        -o "$binary" \
        ./cmd/cdt

    if [[ "$os" == "windows" ]]; then
        (
            cd "$staging"
            zip -q "$dist_abs/$asset.zip" "$(basename "$binary")"
        )
    else
        tar -C "$staging" -czf "$dist_abs/$asset.tar.gz" "$(basename "$binary")"
    fi
    rm -rf "$staging"
    trap - EXIT
done

(
    cd "$dist"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum -- *.tar.gz *.zip > checksums.txt
    else
        shasum -a 256 -- *.tar.gz *.zip > checksums.txt
    fi
)
