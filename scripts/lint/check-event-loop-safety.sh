#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

tmp_file="$(mktemp)"
trap 'rm -f "$tmp_file"' EXIT

if ! rg --files internal/app -g '*.go' -g '!**/*_test.go' >/dev/null; then
	echo "event-loop safety lint: no app go files found"
	exit 0
fi

# shellcheck disable=SC2016
rg --files internal/app -g '*.go' -g '!**/*_test.go' -0 | xargs -0 awk '
function count_braces(s,   i, c) {
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
	trimmed = line
	sub(/^[[:space:]]+/, "", trimmed)

	if (in_target == 0 && line ~ /^func[[:space:]]*\(.*\)[[:space:]]*(Update|View)\(/) {
		in_target = 1
		func_sig = line
		depth = count_braces(line)
	} else if (in_target == 0 && line ~ /^func[[:space:]]+(Update|View)\(/) {
		in_target = 1
		func_sig = line
		depth = count_braces(line)
	} else if (in_target == 1) {
		depth += count_braces(line)
	}

	if (in_target == 1) {
		if (trimmed !~ /^\/\// && line ~ /(context\.WithTimeout\(|time\.Sleep\(|http\.(Get|Post|Do)\(|net\/http|os\.(ReadFile|WriteFile)\(|os\/exec\.|exec\.Command\()/) {
			printf "%s:%d: blocking/external call inside Update/View: %s\n", FILENAME, NR, trimmed
		}
		if (depth <= 0) {
			in_target = 0
			func_sig = ""
			depth = 0
		}
	}
}
' > "$tmp_file"

if [ -s "$tmp_file" ]; then
	echo "event-loop safety lint failed:"
	echo "  Update/View must not perform blocking I/O or external work directly."
	cat "$tmp_file"
	exit 1
fi

echo "event-loop safety lint passed"
