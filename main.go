package main

import (
	"autodoc/api"
	_ "autodoc/docs"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"strings"
)

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}

func enableCORS(router *mux.Router) {
	router.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

					if r.Method == "OPTIONS" {
						w.WriteHeader(http.StatusOK)
						return
					}
					next.ServeHTTP(w, r)
				},
			)
		},
	)
}

// @title AutoDoc API
// @version 1.0
// @description API для AutoDoc
// @BasePath /api/v1
func main() {
	r := mux.NewRouter()
	router := api.Router()
	enableCORS(r)

	mount(r, "/api/v1", router)

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	addr := ":9090"
	log.Printf("🚀 Server started on %s", addr)
	log.Printf("📚 Swagger UI: http://localhost:9090/swagger/index.html")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
