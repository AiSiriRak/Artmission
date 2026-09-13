#!/usr/bin/env bash
# Compare this tree's SQL migrations to a base ref.
# Already-shipped files cannot be rewritten. New files must sort after every
# existing one so goose never sees a missing/out-of-order version on prod.
set -euo pipefail
export LC_ALL=C

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"
MIG_PREFIX="backend/internal/pkg/migrations"
MIG_ABS="$ROOT/$MIG_PREFIX"

BASE_REF="${1:-}"
if [[ -z "$BASE_REF" ]]; then
	if git rev-parse --verify --quiet origin/main >/dev/null; then
		BASE_REF="origin/main"
	else
		echo "usage: $0 <base-ref>" >&2
		exit 2
	fi
fi

if ! git cat-file -e "${BASE_REF}^{commit}" 2>/dev/null; then
	echo "error: cannot resolve base ref '${BASE_REF}'" >&2
	echo "fetch it first, e.g. git fetch origin main" >&2
	exit 2
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

git ls-tree -r --name-only --full-tree "$BASE_REF" -- "$MIG_PREFIX" \
	| awk -F/ '/\.sql$/ { print $NF }' \
	| sort >"$tmp/base"

if [[ -d "$MIG_ABS" ]]; then
	find "$MIG_ABS" -maxdepth 1 -name '*.sql' -exec basename {} \; | sort >"$tmp/head"
else
	: >"$tmp/head"
fi

fail=0

while IFS= read -r name; do
	[[ -z "$name" ]] && continue
	path="$MIG_ABS/$name"
	if [[ ! -f "$path" ]]; then
		continue
	fi
	if ! git show "$BASE_REF:$MIG_PREFIX/$name" | cmp -s - "$path"; then
		echo "error: in-place edit of $MIG_PREFIX/$name" >&2
		echo "       already-applied migrations cannot be rewritten; add a new file." >&2
		fail=1
	fi
done <"$tmp/base"

latest_base="$(tail -n 1 "$tmp/base" || true)"
while IFS= read -r name; do
	[[ -z "$name" ]] && continue
	if [[ -z "$latest_base" ]]; then
		continue
	fi
	last="$(printf '%s\n%s\n' "$latest_base" "$name" | sort | tail -n 1)"
	if [[ "$last" != "$name" || "$name" == "$latest_base" ]]; then
		echo "error: new migration $name is not after latest existing $latest_base" >&2
		echo "       goose applies by version order; inserting in the middle breaks prod." >&2
		fail=1
	fi
done < <(comm -13 "$tmp/base" "$tmp/head")

if [[ "$fail" -ne 0 ]]; then
	exit 1
fi

echo "migration history ok (base ${BASE_REF})"
