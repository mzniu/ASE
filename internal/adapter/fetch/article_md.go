package fetch

import (
	"bytes"
	"net/url"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/go-shiori/go-readability"
)

// HTMLMainToMarkdown extracts readable article HTML with Mozilla Readability, then converts to Markdown.
// On failure, falls back to tag-stripped plain text (still valid in a Markdown document).
func HTMLMainToMarkdown(htmlBytes []byte, pageURL string) string {
	pu, err := url.Parse(pageURL)
	if err != nil {
		pu = nil
	}
	art, err := readability.FromReader(bytes.NewReader(htmlBytes), pu)
	if err != nil || strings.TrimSpace(art.Content) == "" {
		return fallbackPlainFromHTML(string(htmlBytes))
	}
	opts := []converter.ConvertOptionFunc{}
	if pu != nil && pu.Scheme != "" && pu.Host != "" {
		opts = append(opts, converter.WithDomain(pu.Scheme+"://"+pu.Host))
	}
	md, err := htmltomarkdown.ConvertString(art.Content, opts...)
	if err != nil {
		t := strings.TrimSpace(art.TextContent)
		if t != "" {
			return t
		}
		return fallbackPlainFromHTML(string(htmlBytes))
	}
	md = strings.TrimSpace(md)
	if md == "" {
		return fallbackPlainFromHTML(string(htmlBytes))
	}
	return md
}

func fallbackPlainFromHTML(html string) string {
	t := HTMLToPlain(html)
	return strings.TrimSpace(t)
}
