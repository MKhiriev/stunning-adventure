package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// CheckHTTPMethod метод для возвращения статус кода 404 в случае обращения к зарегистрированному роуту с неправильным методом
func CheckHTTPMethod(router *chi.Mux) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestedURL := r.URL.Path
		requestedHTTPMethod := r.Method

		allRoutes := router.Routes()
		var foundRoute chi.Route
		for _, route := range allRoutes {
			if route.Pattern == requestedURL {
				foundRoute = route
				break
			}
		}

		if _, ok := foundRoute.Handlers[requestedHTTPMethod]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		router.ServeHTTP(w, r)
	}
}
