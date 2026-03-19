#!/bin/bash
# Generate per-package line coverage report from coverage.out
# Usage: ./scripts/coverage_by_pkg.sh [coverage.out]
# Output: Sorted list of packages by coverage percentage (lowest first)
#
# This uses the coveragecheck utility to calculate Codecov-compatible
# line coverage, which counts lines that are fully covered.

set -euo pipefail

PROFILE="${1:-coverage.out}"

if [[ ! -f "$PROFILE" ]]; then
    echo "Error: Coverage profile not found: $PROFILE" >&2
    exit 1
fi

go run ./cmd/coveragecheck/by_pkg.go "$PROFILE"
