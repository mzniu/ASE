package baidubrowser

import (
	"testing"
)

func TestParseBaiduContentLeft_sample(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><div id="content_left">
<div class="c-container">
  <h3 class="t"><a href="https://example.com/page">Example title</a></h3>
  <div class="c-abstract"><span class="content">First snippet line.</span></div>
</div>
<div class="c-container">
  <h3><a href="http://www.baidu.com/link?url=enc">Another</a></h3>
  <div class="c-abstract">Second abstract.</div>
</div>
</div></body></html>`
	items := parseBaiduContentLeft(html, 10)
	if len(items) != 2 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0].Title != "Example title" || items[0].Snippet == "" {
		t.Fatalf("%+v", items[0])
	}
	if items[1].Snippet != "Second abstract." {
		t.Fatalf("%+v", items[1])
	}
}

func TestParseBaiduContentLeft_richSnippet(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><div id="content_left">
<div class="c-container">
  <h3><a href="https://example.org/a">Rich title</a></h3>
  <div class="c-abstract"><span>First paragraph of abstract.</span> <span>Second span.</span></div>
  <div class="cosc-content">Supplementary cosco line.</div>
</div>
</div></body></html>`
	items := parseBaiduContentLeft(html, 10)
	if len(items) != 1 {
		t.Fatalf("len=%d", len(items))
	}
	want := "First paragraph of abstract. Second span.\n\nSupplementary cosco line."
	if items[0].Snippet != want {
		t.Fatalf("snippet=%q want %q", items[0].Snippet, want)
	}
}

func TestParseBaiduContentLeft_empty(t *testing.T) {
	if parseBaiduContentLeft("", 5) != nil {
		t.Fatal("expected nil")
	}
}
