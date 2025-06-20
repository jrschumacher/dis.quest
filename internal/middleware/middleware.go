package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	"github.com/jrschumacher/dis.quest/components"
	"github.com/jrschumacher/dis.quest/internal/auth"
)

// AuthMiddleware checks for a valid session cookie and redirects to /login if missing
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := auth.GetSessionCookie(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LayoutMiddleware returns a middleware that wraps the handler's HTML output in components.Page
func LayoutMiddleware(appEnv string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := httptest.NewRecorder()
			next.ServeHTTP(rw, r)

			content := templ.ComponentFunc(func(_ context.Context, wtr io.Writer) error {
				_, err := wtr.Write(rw.Body.Bytes())
				return err
			})

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := components.Page(appEnv, content).Render(r.Context(), w); err != nil {
				http.Error(w, "Failed to render page", http.StatusInternalServerError)
			}
		})
	}
}
