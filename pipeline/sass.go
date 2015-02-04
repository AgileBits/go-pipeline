package pipeline

import (
	"log"
	"net/http"
	"time"
)

func (p *Asset) Sass(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println(">", r.URL)
		h.ServeHTTP(w, r)
		log.Println("<", r.URL, "(", time.Since(start))
	})
}
