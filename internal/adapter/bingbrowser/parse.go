package bingbrowser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/example/ase/internal/port"
)

func collapseWS(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func extractBingSnippet(card *goquery.Selection) string {
	var parts []string
	for _, sel := range []string{".b_caption p", ".b_algoSlug", ".b_snippetBigText", ".b_lineclamp2", ".b_lineclamp3", "p"} {
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
	// De-dupe identical lines
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

// parseBingResults parses desktop Bing SERP HTML for #b_results (best-effort; markup changes).
func parseBingResults(html string, max int) []port.ProviderItem {
	if max < 1 {
		max = 10
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	root := doc.Find("#b_results").First()
	if root.Length() == 0 {
		return nil
	}
	var out []port.ProviderItem
	// Primary: li.b_algo; fallback li with h2
	root.Find("li.b_algo").Each(func(i int, s *goquery.Selection) {
		if len(out) >= max {
			return
		}
		appendBingCard(&out, s, max)
	})
	if len(out) == 0 {
		root.Find("li").Each(func(i int, s *goquery.Selection) {
			if len(out) >= max {
				return
			}
			if s.Find("h2 a").Length() == 0 {
				return
			}
			appendBingCard(&out, s, max)
		})
	}
	return out
}

func appendBingCard(out *[]port.ProviderItem, s *goquery.Selection, max int) {
	if len(*out) >= max {
		return
	}
	a := s.Find("h2 a").First()
	if a.Length() == 0 {
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
	title := strings.TrimSpace(a.Text())
	snippet := extractBingSnippet(s)
	*out = append(*out, port.ProviderItem{
		URL:     href,
		Title:   title,
		Snippet: snippet,
	})
}
