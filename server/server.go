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

// Server contains the methods for serving web requests.
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

func ip(r *http.Request) net.IP {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return net.ParseIP(xff)
	}
	return net.ParseIP(r.RemoteAddr)
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

// NewServer returns a new instance of Server.
func NewServer(r *visitorcounter.Renderer) *Server {
	return &Server{renderer: r}
}

func (s *Server) serveImage(w http.ResponseWriter, r *http.Request) {
	rIP, rDomain := ip(r), domain(r)
	log.Printf("Recieved request from IP: %s [Referer: %q]\n", rIP, rDomain)
	w.Header().Set("Content-Type", "image/png")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := s.renderer.Add(ctx, rIP, rDomain); err != nil {
		log.Print("Server: ", err)
	}
	count := s.renderer.Count(ctx, rDomain)
	img := s.renderer.Render(ctx, parseOptions(r), count)
	png.Encode(w, img)
}

// ListenAndServe listens on the TCP network address addr and serves /c.png
// as well as the files located in ./web at root.
func (s *Server) ListenAndServe(addr string) error {
	http.HandleFunc("/c.png", s.serveImage)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	return http.ListenAndServe(addr, nil)
}
