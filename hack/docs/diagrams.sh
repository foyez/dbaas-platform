#!/usr/bin/env bash
# Renders every view in docs/architecture/workspace.dsl to an SVG in
# docs/architecture/diagrams/: Structurizr exports C4-PlantUML, PlantUML draws it.
# With --check, fails if the committed SVGs don't match the model.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
STRUCTURIZR_IMAGE="structurizr/structurizr:${STRUCTURIZR_VERSION:?run via make diagrams}"
PLANTUML_IMAGE="plantuml/plantuml:${PLANTUML_VERSION:?run via make diagrams}"
SRC="$ROOT/docs/architecture"
OUT="$SRC/diagrams"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
# mktemp -d is 0700 and both images run as their own non-root user, so run them
# as the caller. Docker Desktop and OrbStack hide this on macOS; Linux CI doesn't.
as_me=(--user "$(id -u):$(id -g)")

docker run --rm "${as_me[@]}" -v "$SRC":/src:ro -v "$work":/out "$STRUCTURIZR_IMAGE" \
  export -w /src/workspace.dsl -f plantuml/c4plantuml -o /out >/dev/null

for f in "$work"/structurizr-*.puml; do
  # The exporter writes SHOW_LEGEND(false), which in C4-PlantUML still draws a
  # legend; it would list every tag combination, so drop it.
  sed '/^SHOW_LEGEND/d' "$f" >"$work/$(basename "${f#"$work"/structurizr-}")"
  rm "$f"
done

docker run --rm "${as_me[@]}" -v "$work":/d -w /d "$PLANTUML_IMAGE" -tsvg -failfast2 '*.puml'

if [[ "${1:-}" == "--check" ]]; then
  stale=0
  for f in "$work"/*.svg; do
    if ! cmp -s "$f" "$OUT/$(basename "$f")"; then
      echo "  stale  docs/architecture/diagrams/$(basename "$f")"
      stale=1
    fi
  done
  if [[ $stale -eq 1 ]]; then
    echo "Diagrams are out of date with workspace.dsl. Run 'make diagrams' and commit the result."
    exit 1
  fi
  echo "Diagrams match workspace.dsl."
else
  rm -rf "$OUT"
  mkdir -p "$OUT"
  cp "$work"/*.svg "$OUT"/
  echo "Rendered $(ls "$OUT" | wc -l | tr -d ' ') diagrams into docs/architecture/diagrams/"
fi
