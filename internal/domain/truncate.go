package domain

import "unicode/utf8"

const truncatedNotice = "\n\n（以下内容因长度限制已截断）"

// TruncateToRunes hard-cuts s to maxRunes UTF-8 runes and appends a fixed notice if truncated (REQ-F-009).
func TruncateToRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return truncatedNotice
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	if len(runes) > maxRunes {
		runes = runes[:maxRunes]
	}
	return string(runes) + truncatedNotice
}
