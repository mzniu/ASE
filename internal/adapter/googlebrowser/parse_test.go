package googlebrowser

import (
	"strings"
	"testing"
)

func TestParseGoogleResults_sample(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><div id="center_col">
<div id="rso">
<div class="g"><div class="tF2Cxc">
  <div class="yuRUbf"><a href="https://example.com/page"><h3>Example title</h3></a></div>
  <div class="VwiC3b"><span>Snippet text for result.</span></div>
</div></div>
<div class="g"><div class="tF2Cxc">
  <div class="yuRUbf"><a href="https://other.test/x"><h3>Second</h3></a></div>
</div></div>
</div>
</div></body></html>`
	items := parseGoogleResults(html, 10)
	if len(items) != 2 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0].Title != "Example title" || !strings.Contains(items[0].Snippet, "Snippet text") {
		t.Fatalf("%+v", items[0])
	}
	if items[1].URL != "https://other.test/x" {
		t.Fatalf("%+v", items[1])
	}
}
