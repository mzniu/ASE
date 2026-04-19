package bingbrowser

import (
	"strings"
	"testing"
)

func TestParseBingResults_sample(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><ol id="b_results">
<li class="b_algo"><h2><a href="https://example.com/page">Example title</a></h2>
<div class="b_caption"><p>First line of snippet.</p></div></li>
<li class="b_algo"><h2><a href="https://other.test/x">Other</a></h2>
<p class="b_lineclamp3">Second snippet text.</p></li>
</ol></body></html>`
	items := parseBingResults(html, 10)
	if len(items) != 2 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0].Title != "Example title" || !strings.Contains(items[0].Snippet, "First line") {
		t.Fatalf("%+v", items[0])
	}
	if items[1].URL != "https://other.test/x" {
		t.Fatalf("%+v", items[1])
	}
}
