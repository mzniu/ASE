// Command server is the HTTP entrypoint for the ASE search API.
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/example/ase/internal/adapter/baidubrowser"
	"github.com/example/ase/internal/adapter/bingbrowser"
	"github.com/example/ase/internal/adapter/fetch"
	"github.com/example/ase/internal/adapter/googlebrowser"
	"github.com/example/ase/internal/adapter/opensearch"
	"github.com/example/ase/internal/adapter/stubprovider"
	"github.com/example/ase/internal/adapter/tavily"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/handler"
	apimw "github.com/example/ase/internal/middleware"
	"github.com/example/ase/internal/orchestrator"
	"github.com/example/ase/internal/port"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Optional: .env in the current working directory (repo root when using `go run` from there).
	// godotenv.Load does NOT override keys already present in the process environment; many machines have a stale
	// or empty TAVILY_API_KEY in User/system env, which blocks .env. When .env exists, use Overload so file wins.
	loadDotEnv()

	cfg := config.Load()

	oi, err := opensearch.NewFromConfig(cfg)
	if err != nil {
		log.Fatalf("OpenSearch: %v", err)
	}
	idx := oi
	if _, ok := oi.(*opensearch.Repo); ok {
		log.Printf("IndexRepository: OpenSearch index %q", cfg.OpenSearchIndex)
	} else {
		log.Print("IndexRepository: noop (set OPENSEARCH_URLS + OPENSEARCH_INDEX for OpenSearch)")
	}

	registry := buildProviderRegistry(cfg)
	registry["stub"] = stubprovider.Fixed{
		Result: port.ProviderResult{
			Items: []port.ProviderItem{
				{Title: "stub", Snippet: "将 providers 设为 [\"stub\"] 用于测试；生产请使用 baidu、bing、google 或 tavily。"},
			},
		},
	}
	defaults := effectiveDefaultProviders(cfg, registry)
	log.Printf("SearchProvider registry: %v | default providers: %v", registryKeys(registry), defaults)

	var pageFetch port.PageFetcher = fetch.Noop{}
	if cfg.ProviderFetchResultURLs {
		pageFetch = fetch.NewSimple(fetch.SimpleConfig{
			PerURLTimeout: time.Duration(cfg.FetchPerURLTimeoutMs) * time.Millisecond,
			Concurrency:   cfg.FetchConcurrency,
		})
		log.Print("PageFetcher: result URL fetch enabled (http/https → Markdown excerpts, PROVIDER_FETCH_RESULT_URLS=true)")
	}
	orch := &orchestrator.Service{
		Index:        idx,
		Registry:     registry,
		DefaultNames: defaults,
		Fetcher:      pageFetch,
		Config:       cfg,
	}
	h := handler.NewSearch(cfg, orch)
	docH := handler.NewDocuments(cfg, idx)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", handler.Root)
	r.Get("/api/info", handler.ServiceInfo)
	r.Get("/health", handler.Health)
	r.Handle("/metrics", promhttp.Handler())

	// Agent skill files (embedded; no GitHub) — same auth model as /health
	r.Get("/skills/ase-search-api/SKILL.md", handler.SkillSKILLMD)
	r.Get("/skills/ase-search-api/reference.md", handler.SkillReferenceMD)
	r.Get("/skills/ase-search-api/bundle.zip", handler.SkillBundleZIP)

	r.Group(func(r chi.Router) {
		r.Use(apimw.RateLimit(cfg))
		r.Route("/v1", func(r chi.Router) {
			r.Post("/search", h.Handle)
			r.Post("/documents", docH.Handle)
		})
	})

	addr := cfg.HTTPAddr
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("http server shutdown", "err", err)
	}
}

func buildProviderRegistry(cfg config.Config) map[string]port.SearchProvider {
	r := make(map[string]port.SearchProvider)
	if b := baidubrowser.NewFromConfig(cfg); b != nil {
		r["baidu"] = b
	}
	if b := bingbrowser.NewFromConfig(cfg); b != nil {
		r["bing"] = b
	}
	if g := googlebrowser.NewFromConfig(cfg); g != nil {
		r["google"] = g
	}
	if tc := tavily.NewFromConfig(cfg); tc != nil {
		r["tavily"] = tc
	}
	return r
}

func effectiveDefaultProviders(cfg config.Config, reg map[string]port.SearchProvider) []string {
	if len(cfg.SearchDefaultProviders) > 0 {
		var out []string
		for _, n := range cfg.SearchDefaultProviders {
			n = strings.ToLower(strings.TrimSpace(n))
			if n == "" {
				continue
			}
			if _, ok := reg[n]; ok {
				out = append(out, n)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	order := []string{"baidu", "bing", "google", "tavily"}
	for _, name := range order {
		if _, ok := reg[name]; ok {
			return []string{name}
		}
	}
	return []string{"stub"}
}

func registryKeys(reg map[string]port.SearchProvider) []string {
	keys := make([]string, 0, len(reg))
	for k := range reg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// loadDotEnv applies .env from the process working directory. Uses Overload so keys in .env win over
// pre-existing OS/User environment variables (godotenv.Load does not override existing keys).
func loadDotEnv() {
	const name = ".env"
	st, err := os.Stat(name)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("dotenv: stat %q: %v", name, err)
		}
		return
	}
	if st.IsDir() {
		return
	}
	if err := godotenv.Overload(name); err != nil {
		log.Printf("dotenv: Overload %q: %v", name, err)
		return
	}
	wd, _ := os.Getwd()
	log.Printf("dotenv: loaded %q (cwd=%s); .env overrides existing env for the same keys", filepath.Clean(name), wd)
}
