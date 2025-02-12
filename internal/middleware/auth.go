package middleware

import (
	"fmt"
	"net/http"

	"github.com/kbannyi/gophermart/internal/auth"
	"github.com/kbannyi/gophermart/internal/logger"
)

func AuthExtractor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(auth.HeaderName)
		if token == "" {
			c, err := r.Cookie(auth.CookieName)
			if err == nil {
				token = c.Value
				logger.Log.Debug("Auth cookie found")
			}
		} else {
			logger.Log.Debug("Auth header found")
		}

		if token == "" {
			logger.Log.Debug("Auth token not found")
			h.ServeHTTP(w, r)
			return
		}

		user, err := auth.ReadJWTString(token)
		if err != nil {
			logger.Log.Debug(fmt.Sprintf("Invalid token: %v", err))
			h.ServeHTTP(w, r)
			return
		}
		ctx := auth.ToContext(r.Context(), user)
		logger.Log.Debug("User authenticated", "userid", user.UserID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthGuard(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := auth.FromContext(r.Context())
		if err != nil {
			logger.Log.Debug(fmt.Sprintf("AuthGuard: %v", err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
