package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// longArticleHTML is enough text for Readability to treat the block as main content.
var longArticleHTML = `<!DOCTYPE html><html><head><title>Test</title></head><body>
<article><p>Hello <b>world</b>. ` + strings.Repeat("More readable text here. ", 40) + `</p></article>
</body></html>`

func TestSimple_FetchPlainText_https(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(longArticleHTML))
	}))
	t.Cleanup(ts.Close)

	f := NewSimple(SimpleConfig{PerURLTimeout: 5 * time.Second})
	f.client = ts.Client()

	pages := f.FetchPlainText(context.Background(), []string{ts.URL}, 2)
	if len(pages) != 1 {
		t.Fatalf("pages = %#v", pages)
	}
	plain := strings.ReplaceAll(pages[0].Text, "*", "")
	if !strings.Contains(plain, "Hello") || !strings.Contains(plain, "world") {
		t.Fatalf("text = %q", pages[0].Text)
	}
}

func TestSimple_FetchPlainText_preservesOrderConcurrent(t *testing.T) {
	mux := http.NewServeMux()
	handler := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(body))
		}
	}
	mux.Handle("/a", handler(`<!DOCTYPE html><html><body><article><p>`+strings.Repeat("Alpha page content. ", 50)+`</p></article></body></html>`))
	mux.Handle("/b", handler(`<!DOCTYPE html><html><body><article><p>`+strings.Repeat("Beta page content. ", 50)+`</p></article></body></html>`))
	ts := httptest.NewTLSServer(mux)
	t.Cleanup(ts.Close)

	f := NewSimple(SimpleConfig{PerURLTimeout: 10 * time.Second, Concurrency: 4})
	f.client = ts.Client()

	urlA := ts.URL + "/a"
	urlB := ts.URL + "/b"
	pages := f.FetchPlainText(context.Background(), []string{urlA, urlB}, 2)
	if len(pages) != 2 {
		t.Fatalf("pages = %#v", pages)
	}
	if pages[0].URL != urlA || pages[1].URL != urlB {
		t.Fatalf("order: %#v", pages)
	}
	if !strings.Contains(pages[0].Text, "Alpha") || !strings.Contains(pages[1].Text, "Beta") {
		t.Fatalf("got %#v / %#v", pages[0].Text, pages[1].Text)
	}
}

func TestSimple_skipsNonHTTPSchemes(t *testing.T) {
	f := NewSimple(SimpleConfig{PerURLTimeout: time.Second})
	pages := f.FetchPlainText(context.Background(), []string{"ftp://example.com/page", "javascript:void(0)"}, 2)
	if len(pages) != 0 {
		t.Fatalf("expected skip, got %#v", pages)
	}
}

func TestSimple_FetchPlainText_httpScheme(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(longArticleHTML))
	}))
	t.Cleanup(ts.Close)

	f := NewSimple(SimpleConfig{PerURLTimeout: 5 * time.Second})
	f.client = ts.Client()
	pages := f.FetchPlainText(context.Background(), []string{ts.URL}, 1)
	if len(pages) != 1 || pages[0].URL != ts.URL {
		t.Fatalf("pages = %#v", pages)
	}
}

func TestHTMLToPlain(t *testing.T) {
	in := `<script>x</script><style>.a{}</style><p>Hi &amp; there</p>`
	got := HTMLToPlain(in)
	if got != "Hi & there" {
		t.Fatalf("got %q", got)
	}
}
