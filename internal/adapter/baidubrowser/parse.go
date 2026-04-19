package baidubrowser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/example/ase/internal/port"
)

func collapseWS(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// addPart appends a snippet fragment, dropping redundant substrings so merged cards stay readable.
func addPart(parts *[]string, t string) {
	t = collapseWS(t)
	if t == "" {
		return
	}
	for _, p := range *parts {
		if p == t || strings.Contains(p, t) {
			return
		}
	}
	var next []string
	for _, p := range *parts {
		if strings.Contains(t, p) {
			continue
		}
		next = append(next, p)
	}
	*parts = append(next, t)
}

// extractSnippetFromCard merges typical Baidu organic snippet nodes (DOM varies by template).
func extractSnippetFromCard(card *goquery.Selection) string {
	var parts []string
	for _, sel := range []string{".c-abstract", ".cosc-content", ".cosc-abstract", ".c-span9"} {
		card.Find(sel).Each(func(_ int, n *goquery.Selection) {
			addPart(&parts, n.Text())
		})
	}
	return strings.Join(parts, "\n\n")
}

// parseBaiduContentLeft parses desktop SERP HTML for #content_left (best-effort; Baidu markup changes).
func parseBaiduContentLeft(html string, max int) []port.ProviderItem {
	if max < 1 {
		max = 10
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	left := doc.Find("#content_left").First()
	if left.Length() == 0 {
		return nil
	}
	var out []port.ProviderItem
	// Primary: organic cards; fallback: older .result blocks
	sel := left.Find(".c-container")
	if sel.Length() == 0 {
		sel = left.Find(".result")
	}
	sel.Each(func(i int, s *goquery.Selection) {
		if len(out) >= max {
			return
		}
		a := s.Find("h3 a").First()
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
		snippet := extractSnippetFromCard(s)
		out = append(out, port.ProviderItem{
			URL:     href,
			Title:   title,
			Snippet: snippet,
		})
	})
	return out
}
