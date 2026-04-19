package fetch

import (
	"html"
	"regexp"
	"strings"
)

var (
	reScript = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reStyle  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reTags   = regexp.MustCompile(`<[^>]+>`)
	reWS     = regexp.MustCompile(`\s+`)
)

// HTMLToPlain strips tags and collapses whitespace; best-effort only (no DOM readability).
func HTMLToPlain(s string) string {
	s = reScript.ReplaceAllString(s, " ")
	s = reStyle.ReplaceAllString(s, " ")
	s = reTags.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = strings.TrimSpace(s)
	s = reWS.ReplaceAllString(s, " ")
	return s
}
