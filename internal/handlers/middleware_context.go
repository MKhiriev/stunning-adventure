package handlers

import (
	"context"
	"net/http"
	"time"
)

const timeout = 10 * time.Second

func WithContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFunc := context.WithTimeout(r.Context(), timeout)
		defer cancelFunc()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
