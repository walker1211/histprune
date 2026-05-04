#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")" || exit

echo "Building..."
go build -o histprune ./cmd/histprune
echo "Done. Binary: ./histprune"
