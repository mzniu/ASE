package domain

import (
	"net"
	"net/url"
	"strings"

	"github.com/example/ase/internal/port"
)

const enrichMaxRunes = 8000

// EnrichProviderItemsWithFetch sets BodyMarkdown on items whose URL was fetched successfully.
// Lookup keys are normalized (trim + scheme/host lowercasing) so http/https variants still match.
// Unmatched or failed fetches are ignored; order follows items.
func EnrichProviderItemsWithFetch(items []port.ProviderItem, pages []port.FetchedPage) []port.ProviderItem {
	if len(pages) == 0 {
		return items
	}
	byURL := make(map[string]string, len(pages))
	for _, p := range pages {
		if p.URL == "" || strings.TrimSpace(p.Text) == "" {
			continue
		}
		byURL[FetchURLKey(p.URL)] = p.Text
	}
	if len(byURL) == 0 {
		return items
	}
	out := make([]port.ProviderItem, len(items))
	copy(out, items)
	for i := range out {
		k := FetchURLKey(out[i].URL)
		if k == "" {
			continue
		}
		t, ok := byURL[k]
		if !ok {
			continue
		}
		out[i].BodyMarkdown = TruncateToRunes(strings.TrimSpace(t), enrichMaxRunes)
	}
	return out
}

// FetchURLKey normalizes a URL string for matching fetch results to provider items.
func FetchURLKey(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	u.Scheme = strings.ToLower(u.Scheme)
	if h := u.Hostname(); h != "" {
		port := u.Port()
		if port == "" {
			u.Host = strings.ToLower(h)
		} else {
			u.Host = net.JoinHostPort(strings.ToLower(h), port)
		}
	}
	return u.String()
}
