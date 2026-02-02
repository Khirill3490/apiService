package http

import (
	"context"
	"net/http"
	"strings"

	"api-project/internal/auth"
)

type ctxKey string

const ctxKeyUserID ctxKey = "user_id"

func (h *Handler) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ждём заголовок: Authorization: Bearer <token>
		hdr := r.Header.Get("Authorization")
		if hdr == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(hdr, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ParseAccessToken(tokenString, h.jwtSecret)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Кладём userID в контекст запроса, чтобы handlers могли его достать.
		ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(ctxKeyUserID)
	id, ok := v.(int64)
	return id, ok
}
