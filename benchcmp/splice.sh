#!/usr/bin/env bash
# Usage: ./splice.sh '<TypeRegex>'   e.g. ./splice.sh 'PublicKey|Signature'
# Filters results.txt to the given types, regenerates the markdown section and
# splices it into the repo README between the BENCHMARKS markers.
set -euo pipefail
cd "$(dirname "$0")"

README=../README.md

grep -E "^Benchmark($1)_" results.txt > filtered.txt
go run ./gen filtered.txt > section.md

python3 - "$README" section.md <<'EOF'
import sys
readme_path, section_path = sys.argv[1], sys.argv[2]
readme = open(readme_path).read()
section = open(section_path).read()
begin, end = "<!-- BENCHMARKS:BEGIN -->", "<!-- BENCHMARKS:END -->"
head, _, rest = readme.partition(begin)
_, _, tail = rest.partition(end)
open(readme_path, "w").write(head + begin + "\n" + section + end + tail)
EOF
echo "spliced $1 into README"
