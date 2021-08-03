package server

import (
	"embed"
	"fmt"
	"image/png"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/michaelmcallister/visitorcounter/visitorcounter"
)

const (
	themeParam  = "t"
	widthParam  = "w"
	domainParam = "d"
)

//go:embed web
var content embed.FS

var themeMap = map[string]visitorcounter.Theme{
	"1": visitorcounter.Segment,
	"2": visitorcounter.Aomm,
}

// Server contains the methods for serving web requests.
type Server struct {
	renderer *visitorcounter.Renderer
	server   *http.Server
}

type webroot struct {
	http.FileSystem
}

func (wr webroot) Open(name string) (http.File, error) {
	n := fmt.Sprintf("web%s", name)
	if name == "/" {
		n = "web"
	}
	return wr.FileSystem.Open(n)
}

func domain(r *http.Request) string {
	return r.URL.Query().Get(domainParam)
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
	log.Printf("Received request from IP: %s [Referer: %q]\n", rIP, rDomain)

	if err := s.renderer.Add(rIP, rDomain); err != nil {
		// We continue on error when trying to add a domain to the datastore,
		// as it's not the end of the world.
		log.Print("Server: ", err)
	}
	count := s.renderer.Count(rDomain)
	img, err := s.renderer.Render(parseOptions(r), count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print("Server: ", err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

// ListenAndServe listens on the TCP network address addr and serves /c.png
// as well as the files located in ./web at root.
func (s *Server) ListenAndServe(addr string) error {
	http.HandleFunc("/c.png", s.serveImage)

	http.Handle("/", http.FileServer(webroot{http.FS(content)}))
	return http.ListenAndServe(addr, nil)
}
