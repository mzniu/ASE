package domain

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	reMDLink  = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	reMDBold  = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reFence   = regexp.MustCompile("(?s)```[^`]*```")
	reHeading = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	reHTMLTag = regexp.MustCompile(`<[^>]+>`)
	reMultiNL = regexp.MustCompile(`\n{3,}`)
)

// WritebackBodyFromMarkdown converts provider-path Markdown into plain-ish text for OpenSearch body_text (Phase-1 index write-back).
func WritebackBodyFromMarkdown(md string, maxRunes int) string {
	s := strings.TrimSpace(md)
	if s == "" {
		return ""
	}
	s = reFence.ReplaceAllString(s, "\n\n")
	s = reMDLink.ReplaceAllString(s, "$1")
	s = reMDBold.ReplaceAllString(s, "$1")
	s = reHeading.ReplaceAllString(s, "")
	s = reHTMLTag.ReplaceAllString(s, "")
	s = reMultiNL.ReplaceAllString(s, "\n\n")
	s = strings.TrimSpace(s)
	return hardTruncateRunes(s, maxRunes)
}

func hardTruncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	r := []rune(s)
	return string(r[:maxRunes])
}

// WritebackTitleFromQuery is the OpenSearch title field for query-keyed write-back documents.
func WritebackTitleFromQuery(query string, maxRunes int) string {
	return hardTruncateRunes(strings.TrimSpace(query), maxRunes)
}
