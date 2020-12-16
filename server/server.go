package server

import (
	"context"
	"image/png"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/michaelmcallister/visitorcounter/visitorcounter"
)

const (
	themeParam  = "t"
	widthParam  = "w"
	domainParam = "d"
)

var themeMap = map[string]visitorcounter.Theme{
	"1": visitorcounter.Segment,
	"2": visitorcounter.Aomm,
}

type Server struct {
	renderer *visitorcounter.Renderer
	server   *http.Server
}

func domain(r *http.Request) string {
	d := r.Header.Get("Referer")
	// Use the Referer header if available. If not, fall back to (optionally)
	// supplied domain parameter.
	if d == "" {
		d = r.URL.Query().Get(domainParam)
	}
	return d
}

func parseOptions(r *http.Request) *visitorcounter.Options {
	tp := r.URL.Query().Get(themeParam)
	theme := visitorcounter.Segment
	if v, ok := themeMap[tp]; ok {
		theme = v
	}

	w := r.URL.Query().Get(widthParam)
	width := visitorcounter.DefaultWidth
	i, err := strconv.Atoi(w)
	if err == nil {
		width = i
	}
	return &visitorcounter.Options{
		Theme: theme,
		Width: width,
	}
}

func NewServer(r *visitorcounter.Renderer) (*Server, error) {
	return &Server{renderer: r}, nil
}

func (s *Server) serveImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")

	log.Printf("Recieved request from IP: %s [Referer: %q]\n", r.RemoteAddr, domain(r))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	go s.renderer.Add(ctx, net.ParseIP(r.RemoteAddr), domain(r))

	count := s.renderer.Count(ctx, domain(r))
	img, err := s.renderer.Render(ctx, parseOptions(r), count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalf("unable to render counter: %v", err)
		return
	}
	png.Encode(w, img)
}

func (s *Server) ListenAndServe() error {
	http.HandleFunc("/c.png", s.serveImage)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	return http.ListenAndServe(":8080", nil)
}
