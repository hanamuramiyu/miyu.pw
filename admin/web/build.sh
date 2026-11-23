#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

BUN="${BUN:-bun}"

"$BUN" install --frozen-lockfile
"$BUN" run build
