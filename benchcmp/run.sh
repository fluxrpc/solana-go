#!/usr/bin/env bash
# Usage: ./run.sh <TypePattern>
# e.g.:  ./run.sh PublicKey
#        ./run.sh '(PublicKey|Signature)'
#
# Runs the flux-vs-gagl benchmarks matching the pattern, appends the raw
# output to results.txt (later runs of the same benchmark override earlier
# ones in the generated markdown), then regenerates readme_section.md.
set -euo pipefail
cd "$(dirname "$0")"

pattern="${1:?usage: ./run.sh <TypePattern>}"

go test -bench="^Benchmark${pattern}_" -benchmem -count=1 -benchtime=1s . | tee -a results.txt
go run ./gen results.txt > readme_section.md
echo "wrote $(pwd)/readme_section.md"
