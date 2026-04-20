package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/example/ase/internal/admin"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/httpx"
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
func RegisterAdmin(r chi.Router, cfg config.Config, signer *admin.SessionSigner) {
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

	r.Post("/admin/api/login", adminLogin(cfg, signer))
	r.Post("/admin/api/logout", adminLogout)
	r.Get("/admin/api/session", adminSession(signer))

	r.Group(func(r chi.Router) {
		r.Use(adminAuthMiddleware(signer))
		r.Get("/admin/api/config", adminConfigView(cfg))
		r.Get("/admin/api/indices", adminIndices(cfg))
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
