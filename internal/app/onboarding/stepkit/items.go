package stepkit

import "github.com/usetero/cli/internal/tea/components/remotelist"

// CastItems converts remotelist items to the target type, discarding mismatches.
func CastItems[T any](items []remotelist.Item) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		v, ok := item.(T)
		if !ok {
			continue
		}
		out = append(out, v)
	}
	return out
}
