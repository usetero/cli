#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

event_dir="${TERO_EVENT_DIR:-internal/app/events}"
allowed_suffix='(Requested|Published|Changed)$'

if ! [ -d "$event_dir" ]; then
	echo "event naming lint: directory not found: ${event_dir}"
	exit 1
fi

mapfile -t event_types < <(rg --no-filename --only-matching --replace '$1' '^type ([A-Z][A-Za-z0-9_]*) struct' "$event_dir" -g '*.go' | sort -u)

if [ "${#event_types[@]}" -eq 0 ]; then
	echo "event naming lint: no exported event structs found in ${event_dir}"
	exit 0
fi

violations=()
for event_type in "${event_types[@]}"; do
	if [[ ! "$event_type" =~ $allowed_suffix ]]; then
		violations+=("$event_type")
	fi
done

if [ "${#violations[@]}" -gt 0 ]; then
	echo "event naming lint failed:"
	echo "  exported event structs in ${event_dir} must end with Requested, Published, or Changed"
	for name in "${violations[@]}"; do
		echo "  - ${name}"
	done
	exit 1
fi

echo "event naming lint passed (${#event_types[@]} event types)"
