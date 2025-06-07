package middleware

import (
	"net/http"

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

