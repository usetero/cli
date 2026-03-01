#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

scan_root="${TERO_SCAN_ROOT:-$repo_root}"
module_root="${TERO_MODULE_ROOT:-$scan_root}"
app_root="${TERO_APP_ROOT:-$scan_root/internal/app}"
module_prefix="${TERO_MODULE_PREFIX:-github.com/usetero/cli}"

if ! [ -d "$app_root" ]; then
	echo "event ownership lint: app root not found: ${app_root}"
	exit 1
fi

violations=()
feature_event_dirs_found=0

while IFS= read -r event_dir; do
	# Skip shared cross-app events package.
	if [ "$event_dir" = "$app_root/events" ]; then
		continue
	fi
	feature_event_dirs_found=$((feature_event_dirs_found + 1))

	owner_dir="${event_dir%/events}"
	rel_event_dir="${event_dir#"$module_root"/}"
	if [ "$rel_event_dir" = "$event_dir" ]; then
		echo "event ownership lint: cannot derive module-relative path for ${event_dir}"
		exit 1
	fi

	import_path="${module_prefix}/${rel_event_dir}"

	while IFS= read -r match; do
		file_path="${match%%:*}"
		# Allow owner feature subtree imports.
		case "$file_path" in
		"$owner_dir"/*) continue ;;
		esac

		# Allow top-level app orchestration files (internal/app/*.go).
		rel_from_app="${file_path#"$app_root"/}"
		if [ "$rel_from_app" != "$file_path" ] && [[ "$rel_from_app" != */* ]]; then
			continue
		fi

		violations+=("${file_path} imports ${import_path} outside owner ${owner_dir}")
	done < <(rg -n --glob '*.go' --glob '!**/*_test.go' "\"${import_path}\"" "$scan_root" || true)
done < <(find "$app_root" -type d -name events | sort)

if [ "${#violations[@]}" -gt 0 ]; then
	echo "event ownership lint failed:"
	echo "  feature event packages may only be imported by their owner feature subtree or internal/app root files."
	for v in "${violations[@]}"; do
		echo "  - ${v}"
	done
	exit 1
fi

if [ "$feature_event_dirs_found" -eq 0 ]; then
	echo "event ownership lint passed (no feature event packages found)"
else
	echo "event ownership lint passed (${feature_event_dirs_found} feature event package(s))"
fi
