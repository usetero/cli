#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

allowed_suffix='(Completed|Loaded|Requested|Tick|Poll|Updated|Created|Validated|Refreshed|Changed|Started|Update)$'

mapfile -t local_msgs < <(rg \
	--no-filename \
	--only-matching \
	--replace '$1' \
	'^type ([a-z][A-Za-z0-9_]*)Msg struct' \
	internal/app \
	-g '*.go' \
	-g '!**/*_test.go' | sort -u)

if [ "${#local_msgs[@]}" -eq 0 ]; then
	echo "local msg naming lint: no local message structs found under internal/app"
	exit 0
fi

violations=()
for msg_name in "${local_msgs[@]}"; do
	if [[ ! "$msg_name" =~ $allowed_suffix ]]; then
		violations+=("${msg_name}Msg")
	fi
done

if [ "${#violations[@]}" -gt 0 ]; then
	echo "local msg naming lint failed:"
	echo "  local message structs must use semantic suffixes before Msg"
	echo "  allowed suffixes: Completed, Loaded, Requested, Tick, Poll, Updated, Created, Validated, Refreshed, Changed, Started, Update"
	for name in "${violations[@]}"; do
		echo "  - ${name}"
	done
	exit 1
fi

echo "local msg naming lint passed (${#local_msgs[@]} local message types)"
