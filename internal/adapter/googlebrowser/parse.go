package googlebrowser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/example/ase/internal/port"
)

func collapseWS(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func extractGoogleSnippet(card *goquery.Selection) string {
	var parts []string
	for _, sel := range []string{
		".VwiC3b", ".yXK7lf .VwiC3b", ".IsZvec", ".lEBKkf",
		".lyLwlc", "span.st", ".aCOpRe", ".MUxGbd",
	} {
		card.Find(sel).Each(func(_ int, n *goquery.Selection) {
			t := collapseWS(n.Text())
			if t == "" {
				return
			}
			parts = append(parts, t)
		})
		if len(parts) > 0 {
			break
		}
	}
	if len(parts) == 0 {
		return ""
	}
	seen := make(map[string]struct{})
	var out []string
	for _, p := range parts {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return strings.Join(out, "\n\n")
}

func findAnchorWithH3(card *goquery.Selection) *goquery.Selection {
	var hit *goquery.Selection
	card.Find("a").Each(func(_ int, s *goquery.Selection) {
		if hit != nil {
			return
		}
		h := s.Find("h3").First()
		if h.Length() == 0 {
			return
		}
		if strings.TrimSpace(h.Text()) != "" {
			hit = s
		}
	})
	return hit
}

// parseGoogleResults parses Google desktop SERP HTML (best-effort; markup changes frequently).
func parseGoogleResults(html string, max int) []port.ProviderItem {
	if max < 1 {
		max = 10
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	root := doc.Find("#center_col").First()
	if root.Length() == 0 {
		root = doc.Find("#search").First()
	}
	if root.Length() == 0 {
		root = doc.Find("#rso").First()
	}
	if root.Length() == 0 {
		return nil
	}
	var out []port.ProviderItem
	root.Find("div.g").Each(func(_ int, s *goquery.Selection) {
		if len(out) >= max {
			return
		}
		appendGoogleCard(&out, s, max)
	})
	return out
}

func appendGoogleCard(out *[]port.ProviderItem, card *goquery.Selection, max int) {
	if len(*out) >= max {
		return
	}
	a := findAnchorWithH3(card)
	if a == nil || a.Length() == 0 {
		return
	}
	href, ok := a.Attr("href")
	if !ok {
		return
	}
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "javascript:") {
		return
	}
	if isGoogleNoiseURL(href) {
		return
	}
	title := strings.TrimSpace(a.Find("h3").First().Text())
	if title == "" {
		return
	}
	snippet := extractGoogleSnippet(card)
	*out = append(*out, port.ProviderItem{
		URL:     href,
		Title:   title,
		Snippet: snippet,
	})
}

func isGoogleNoiseURL(href string) bool {
	low := strings.ToLower(href)
	switch {
	case strings.HasPrefix(low, "#"):
		return true
	case strings.HasPrefix(low, "/search"):
		return true
	case strings.Contains(low, "google.com/maps"):
		return true
	case strings.Contains(low, "google.com/search") && !strings.Contains(low, "/url?"):
		return true
	}
	return false
}
