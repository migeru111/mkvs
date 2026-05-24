#!/bin/bash
set -e

go run ./cmd/bench "$@"

git add reports/
git commit -m "bench: report for $(git rev-parse --short HEAD)"
