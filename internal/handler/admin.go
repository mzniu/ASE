package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/ase/internal/admin"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/httpx"
	"github.com/example/ase/internal/port"
	"github.com/example/ase/internal/webcontent"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const adminCookieName = "ase_admin"

type adminLoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterAdmin mounts /admin UI and APIs when cfg.AdminUIEnabled(); no-op otherwise.
func RegisterAdmin(r chi.Router, cfg config.Config, signer *admin.SessionSigner, idx port.IndexRepository) {
	if !cfg.AdminUIEnabled() || signer == nil {
		return
	}

	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusFound)
	})
	r.Get("/admin/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(webcontent.AdminHTML)
	})
	r.Get("/admin/opensearch", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/opensearch/", http.StatusFound)
	})
	r.Get("/admin/opensearch/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(webcontent.AdminOpenSearchHTML)
	})

	r.Post("/admin/api/login", adminLogin(cfg, signer))
	r.Post("/admin/api/logout", adminLogout)
	r.Get("/admin/api/session", adminSession(signer))

	r.Group(func(r chi.Router) {
		r.Use(adminAuthMiddleware(signer))
		r.Get("/admin/api/config", adminConfigView(cfg))
		r.Get("/admin/api/indices", adminIndices(cfg))
		r.Get("/admin/api/opensearch/meta", adminOpenSearchMeta(cfg))
		r.Get("/admin/api/opensearch/documents", adminOpenSearchDocuments(cfg))
		r.Get("/admin/api/opensearch/hits", adminOpenSearchHits(cfg, idx))
	})
}

func adminLogin(cfg config.Config, signer *admin.SessionSigner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := middleware.GetReqID(r.Context())
		if ct := r.Header.Get("Content-Type"); ct != "" && ct != "application/json" && !strings.HasPrefix(ct, "application/json") {
			httpx.WriteProblem(w, http.StatusUnsupportedMediaType, "unsupported media type", "Content-Type must be application/json")
			return
		}
		var body adminLoginBody
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body); err != nil {
			httpx.WriteProblem(w, http.StatusBadRequest, "invalid json", err.Error())
			return
		}
		if !admin.CheckPassword(cfg, body.Username, body.Password) {
			slog.Warn("admin login failed", "request_id", rid)
			httpx.WriteProblem(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
			return
		}
		tok, err := signer.Issue(time.Now())
		if err != nil {
			slog.Error("admin session issue", "request_id", rid, "err", err)
			httpx.WriteProblem(w, http.StatusInternalServerError, "session error", "could not issue session")
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     adminCookieName,
			Value:    tok,
			Path:     "/admin",
			MaxAge:   int(cfg.AdminSessionTTL.Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   adminCookieSecure(r),
		})
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}
}

func adminCookieSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

func adminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    "",
		Path:     "/admin",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   adminCookieSecure(r),
	})
	w.WriteHeader(http.StatusNoContent)
}

func adminSession(signer *admin.SessionSigner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok := false
		if c, err := r.Cookie(adminCookieName); err == nil && signer.Verify(c.Value, time.Now()) {
			ok = true
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if ok {
			_, _ = w.Write([]byte(`{"ok":true}`))
		} else {
			_, _ = w.Write([]byte(`{"ok":false}`))
		}
	}
}

func adminAuthMiddleware(signer *admin.SessionSigner) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(adminCookieName)
			if err != nil || !signer.Verify(c.Value, time.Now()) {
				httpx.WriteProblem(w, http.StatusUnauthorized, "unauthorized", "admin session required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func adminConfigView(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(admin.ConfigSnapshot(cfg))
	}
}

func adminIndices(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		b, status, err := admin.CatIndicesJSON(ctx, cfg)
		if err != nil {
			slog.Error("admin indices", "err", err)
			httpx.WriteProblem(w, http.StatusBadGateway, "opensearch error", err.Error())
			return
		}
		if status != http.StatusOK {
			httpx.WriteProblem(w, http.StatusBadGateway, "opensearch error", string(b))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}

func adminOpenSearchMeta(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok := len(cfg.OpenSearchURLs) > 0 && strings.TrimSpace(cfg.OpenSearchIndex) != ""
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"open_search_configured": ok,
			"index":                  cfg.OpenSearchIndex,
		})
	}
}

func adminOpenSearchDocuments(cfg config.Config) http.HandlerFunc {
	type browseBody struct {
		Query map[string]any `json:"query"`
		From  int            `json:"from"`
		Size  int            `json:"size"`
		Sort  []string       `json:"sort"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		from := clampInt(queryIntDefault(r, "from", 0), 0, 50000)
		size := clampInt(queryIntDefault(r, "size", 20), 1, 100)
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()
		raw, err := json.Marshal(browseBody{
			Query: map[string]any{"match_all": map[string]any{}},
			From:  from,
			Size:  size,
			Sort:  []string{"_doc"},
		})
		if err != nil {
			httpx.WriteProblem(w, http.StatusInternalServerError, "encode error", err.Error())
			return
		}
		b, status, err := admin.IndexSearchRaw(ctx, cfg, raw)
		if err != nil {
			if status == http.StatusServiceUnavailable {
				httpx.WriteProblem(w, http.StatusServiceUnavailable, "not available", err.Error())
				return
			}
			slog.Error("admin opensearch documents", "err", err)
			httpx.WriteProblem(w, http.StatusBadGateway, "opensearch error", err.Error())
			return
		}
		if status != http.StatusOK {
			httpx.WriteProblem(w, http.StatusBadGateway, "opensearch error", string(b))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}

func adminOpenSearchHits(cfg config.Config, idx port.IndexRepository) http.HandlerFunc {
	type hitRow struct {
		ID    string  `json:"id"`
		Score float64 `json:"score"`
		Body  string  `json:"body"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if q == "" {
			httpx.WriteProblem(w, http.StatusBadRequest, "bad request", "query parameter q is required")
			return
		}
		if len(cfg.OpenSearchURLs) == 0 || strings.TrimSpace(cfg.OpenSearchIndex) == "" {
			httpx.WriteProblem(w, http.StatusServiceUnavailable, "not available", "OpenSearch not configured")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()
		hits, err := idx.Search(ctx, q)
		if err != nil {
			slog.Error("admin opensearch hits", "err", err)
			httpx.WriteProblem(w, http.StatusBadGateway, "search error", err.Error())
			return
		}
		out := make([]hitRow, 0, len(hits))
		for _, h := range hits {
			out = append(out, hitRow{ID: h.ID, Score: h.Score, Body: h.Body})
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query": q,
			"hits":  out,
			"count": len(out),
		})
	}
}

func queryIntDefault(r *http.Request, key string, def int) int {
	s := strings.TrimSpace(r.URL.Query().Get(key))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// RegisterAdminDisabledRoutes registers /admin with 503 + instructions when Admin UI env is not set.
func RegisterAdminDisabledRoutes(r chi.Router) {
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusFound)
	})
	r.Get("/admin/", adminDisabledPage)
}

func adminDisabledPage(w http.ResponseWriter, _ *http.Request) {
	const page = `<!DOCTYPE html><html lang="zh-CN"><head><meta charset="utf-8"/><title>Admin 未启用</title></head><body style="font-family:system-ui,sans-serif;max-width:40rem;margin:2rem;line-height:1.5">
<h1>Admin 管理页未配置</h1>
<p>当前未设置启用 Admin 所需的环境变量，因此无法使用登录与配置页。</p>
<p>请在部署环境（如仓库根目录 <code>.env</code>）中配置并重启 ASE 容器：</p>
<ul>
<li><code>ADMIN_USERNAME</code></li>
<li><code>ADMIN_PASSWORD_BCRYPT</code>（推荐）或开发用 <code>ADMIN_PASSWORD</code></li>
<li><code>ADMIN_SESSION_SECRET</code>（至少 16 个字符）</li>
</ul>
<p>Docker Compose 部署时须将上述变量传入 <code>ase</code> 服务（本仓库 <code>docker-compose.yml</code> 已支持从 <code>.env</code> 注入）。</p>
<p>详见仓库 <code>docs/ADMIN_ENABLE.md</code>。</p>
</body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(page))
}
