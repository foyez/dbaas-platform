#!/usr/bin/env bash
# Renders every Mermaid diagram under docs/ with mermaid-cli, so syntax errors
# fail the build instead of showing up as an error box on the published site.
# `mkdocs build --strict` doesn't catch these: Mermaid renders in the browser.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="minlag/mermaid-cli:${MERMAID_CLI_VERSION:?run via make docs-check}"

cd "$ROOT"
# Plain loop instead of mapfile: macOS still ships bash 3.2.
files=()
while IFS= read -r f; do files+=("$f"); done < <(grep -rl --include='*.md' '^```mermaid' docs | sort)
if [[ ${#files[@]} -eq 0 ]]; then
  echo "No Mermaid diagrams found."
  exit 0
fi
echo "Checking Mermaid diagrams in ${#files[@]} files with $IMAGE"

# One container for all files; mmdc renders every mermaid block in a markdown
# file and exits non-zero on the first parse error.
docker run --rm -v "$ROOT/docs":/docs:ro --entrypoint sh "$IMAGE" -c '
  failed=0
  for f in "$@"; do
    out="/tmp/out/${f#docs/}"
    mkdir -p "$(dirname "$out")"
    if log=$(/home/mermaidcli/node_modules/.bin/mmdc -p /puppeteer-config.json \
        -i "/$f" -o "$out" 2>&1); then
      echo "  ok    $f"
    else
      echo "  FAIL  $f"
      echo "$log" | grep -v "^    at " | sed "s/^/        /"
      failed=1
    fi
  done
  exit $failed
' sh "${files[@]}"
