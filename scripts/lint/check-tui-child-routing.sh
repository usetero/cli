#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

tui_screens_root="${TERO_TUI_SCREENS_ROOT:-internal/interfaces/tui/screens}"

if ! [ -d "$tui_screens_root" ]; then
	echo "tui child routing lint: directory not found: ${tui_screens_root}"
	exit 1
fi

mapfile -t router_files < <(rg -l 'router[[:space:]]+screen\.Router' "$tui_screens_root" -g '*model.go' -g '!**/*_test.go' | sort)

if [ "${#router_files[@]}" -eq 0 ]; then
	echo "tui child routing lint passed (no router-backed models found)"
	exit 0
fi

violations=()

for file in "${router_files[@]}"; do
	if ! rg -q 'm\.router\.Forward\(msg\)' "$file"; then
		violations+=("${file}: missing router.Forward(msg) in Update path")
	fi

	if ! rg -q 'm\.router\.(ActivateOnly|SetActive|ClearActive)\(' "$file"; then
		violations+=("${file}: missing child activation call (ActivateOnly/SetActive/ClearActive)")
	fi

	if ! rg -q 'm\.router\.ShortHelp\(\)' "$file"; then
		violations+=("${file}: missing ShortHelp cascade via router.ShortHelp()")
	fi

	# Keep state mutations in Update; tea.Cmd closures should emit messages only.
	if awk '
function count_braces(s,   i, c, ch) {
	c = 0
	for (i = 1; i <= length(s); i++) {
		ch = substr(s, i, 1)
		if (ch == "{") c++
		if (ch == "}") c--
	}
	return c
}
{
	line = $0
	if (in_cmd == 0 && line ~ /func\(\)[[:space:]]*tea\.Msg[[:space:]]*\{/) {
		in_cmd = 1
		depth = count_braces(line)
	} else if (in_cmd == 1) {
		depth += count_braces(line)
	}

	if (in_cmd == 1) {
		trimmed = line
		sub(/^[[:space:]]+/, "", trimmed)
		if (trimmed !~ /^\/\// &&
			trimmed ~ /m\.[A-Za-z0-9_\.]+[[:space:]]*[\+\-\*\/]?=/ &&
			trimmed !~ /==|!=|>=|<=/) {
			printf "%s:%d: state mutation inside tea.Cmd closure: %s\n", FILENAME, NR, trimmed
			found = 1
		}
		if (trimmed !~ /^\/\// && trimmed ~ /m\.[A-Za-z0-9_\.]+(\+\+|--)/) {
			printf "%s:%d: state mutation inside tea.Cmd closure: %s\n", FILENAME, NR, trimmed
			found = 1
		}
		if (depth <= 0) {
			in_cmd = 0
			depth = 0
		}
	}
}
END {
	if (found == 1) {
		exit 1
	}
}
' "$file"; then
		:
	else
		violations+=("${file}: state mutation found inside tea.Cmd closure")
	fi
done

if [ "${#violations[@]}" -gt 0 ]; then
	echo "tui child routing lint failed:"
	echo "  router-backed parent models must route messages consistently and keep state changes in Update."
	for violation in "${violations[@]}"; do
		echo "  - ${violation}"
	done
	exit 1
fi

echo "tui child routing lint passed (${#router_files[@]} router-backed model files)"
