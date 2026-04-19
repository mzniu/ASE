package domain

import (
	"strconv"
	"strings"

	"github.com/example/ase/internal/port"
)

// MarkdownFromIndex builds non–portal-list style Markdown from index hits (REQ-F-005).
func MarkdownFromIndex(query string, hits []port.Hit) string {
	var b strings.Builder
	b.WriteString("# 检索整理\n\n")
	b.WriteString("**查询**：")
	b.WriteString(query)
	b.WriteString("\n\n")
	b.WriteString("## 要点\n\n")
	n := 0
	for _, h := range hits {
		if strings.TrimSpace(h.Body) == "" {
			continue
		}
		n++
		b.WriteString("### 片段 ")
		b.WriteString(strconv.Itoa(n))
		b.WriteString("\n\n")
		b.WriteString(strings.TrimSpace(h.Body))
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

// MarkdownFromProvider builds Markdown from fallback provider items when index is insufficient.
// SERP snippets go under 「摘要」; full-page fetches (BodyMarkdown) go under 「正文」.
func MarkdownFromProvider(query string, items []port.ProviderItem) string {
	var b strings.Builder
	b.WriteString("# 联网检索（回落）\n\n")
	b.WriteString("**查询**：")
	b.WriteString(query)
	b.WriteString("\n\n")
	if len(items) == 0 {
		b.WriteString("_（无候选结果）_\n")
		return b.String()
	}
	b.WriteString("## 摘要\n\n")
	for i, it := range items {
		title := strings.TrimSpace(it.Title)
		if title == "" {
			title = "结果 " + strconv.Itoa(i+1)
		}
		b.WriteString("### ")
		b.WriteString(title)
		b.WriteString("\n\n")
		if u := strings.TrimSpace(it.URL); u != "" {
			b.WriteString("**链接**：")
			b.WriteString(u)
			b.WriteString("\n\n")
		}
		if src := strings.TrimSpace(it.Source); src != "" {
			b.WriteString("**来源**：")
			b.WriteString(src)
			b.WriteString("\n\n")
		}
		body := strings.TrimSpace(it.Snippet)
		if body != "" && body == strings.TrimSpace(it.Title) {
			body = ""
		}
		if body != "" {
			b.WriteString(body)
			b.WriteString("\n\n")
		}
	}
	hasBody := false
	for _, it := range items {
		if strings.TrimSpace(it.BodyMarkdown) != "" {
			hasBody = true
			break
		}
	}
	if hasBody {
		b.WriteString("## 正文\n\n")
		for i, it := range items {
			md := strings.TrimSpace(it.BodyMarkdown)
			if md == "" {
				continue
			}
			title := strings.TrimSpace(it.Title)
			if title == "" {
				title = "结果 " + strconv.Itoa(i+1)
			}
			b.WriteString("### ")
			b.WriteString(title)
			b.WriteString("\n\n")
			if u := strings.TrimSpace(it.URL); u != "" {
				b.WriteString("**链接**：")
				b.WriteString(u)
				b.WriteString("\n\n")
			}
			if src := strings.TrimSpace(it.Source); src != "" {
				b.WriteString("**来源**：")
				b.WriteString(src)
				b.WriteString("\n\n")
			}
			b.WriteString(md)
			b.WriteString("\n\n")
		}
	}
	return strings.TrimSpace(b.String()) + "\n"
}
