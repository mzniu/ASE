package domain

import (
	"strings"

	"github.com/example/ase/internal/port"
)

// MergeProviderItemsDedupe merges rows that share the same FetchURLKey; snippets are concatenated with a separator.
// Order follows the input order (first occurrence wins position); later duplicates merge into that row.
func MergeProviderItemsDedupe(items []port.ProviderItem) []port.ProviderItem {
	if len(items) <= 1 {
		return items
	}
	keyToIdx := make(map[string]int)
	var out []port.ProviderItem
	for _, it := range items {
		k := FetchURLKey(it.URL)
		if k == "" {
			out = append(out, it)
			continue
		}
		if idx, ok := keyToIdx[k]; ok {
			prev := &out[idx]
			if s := strings.TrimSpace(it.Snippet); s != "" {
				if strings.TrimSpace(prev.Snippet) == "" {
					prev.Snippet = it.Snippet
				} else if !strings.Contains(prev.Snippet, s) {
					src := strings.TrimSpace(it.Source)
					if src != "" {
						prev.Snippet = prev.Snippet + "\n\n— — —\n\n（" + src + "）\n\n" + it.Snippet
					} else {
						prev.Snippet = prev.Snippet + "\n\n— — —\n\n" + it.Snippet
					}
				}
			}
			if it.Source != "" && !sourceSetContains(prev.Source, it.Source) {
				if prev.Source == "" {
					prev.Source = it.Source
				} else {
					prev.Source = prev.Source + ", " + it.Source
				}
			}
			if len(strings.TrimSpace(it.BodyMarkdown)) > len(strings.TrimSpace(prev.BodyMarkdown)) {
				prev.BodyMarkdown = it.BodyMarkdown
			}
			continue
		}
		keyToIdx[k] = len(out)
		out = append(out, it)
	}
	return out
}

func sourceSetContains(commaList, one string) bool {
	if commaList == "" || one == "" {
		return false
	}
	for _, p := range strings.Split(commaList, ",") {
		if strings.TrimSpace(p) == one {
			return true
		}
	}
	return false
}
