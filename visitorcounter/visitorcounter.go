package visitorcounter

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"log"
	"net"
	"time"

	"github.com/michaelmcallister/visitorcounter/datastore"
)

const (
	// DefaultWidth is the minimum padding when displaying numbers, for example
	// rendering the integer '0' with a defaultWidth of 5 would render as
	// '00000'
	DefaultWidth    = 5
	maxWidth        = 100
	saveQueueLength = 100
)

// Theme represents different styles for rendering.
type Theme string

//go:embed theme
var imgdir embed.FS

const (
	themeRootDir = "theme"
)
const (
	// Segment is a classic red 7-segment display.
	Segment Theme = "segment"
	// Aomm is the font from http://aomm.xyz/
	Aomm Theme = "aomm"
)

var themeTiles map[Theme][]image.Image

// Renderer is responsible for queing write events to the datastore, as well
// as retrieving counts and redering them as PNGs.
type Renderer struct {
	datastore datastore.EventWriterCounter
	save      chan datastore.VisitEvent
}

// Options influence how to render the PNG.
type Options struct {
	Theme Theme
	Width int
}

func init() {
	themeTiles = make(map[Theme][]image.Image)
	dirs, err := imgdir.ReadDir(themeRootDir)
	if err != nil {
		log.Fatalln(err)
	}
	for _, dir := range dirs {
		d, err := imgdir.ReadDir(themeRootDir + "/" + dir.Name())
		if err != nil {
			log.Fatalln(err)
		}
		var tiles []image.Image
		for _, file := range d {
			f := fmt.Sprintf("%s/%s/%s", themeRootDir, dir.Name(), file.Name())
			b, err := imgdir.ReadFile(f)
			if err != nil {
				log.Fatalln(err)
			}
			img, _, err := image.Decode(bytes.NewReader(b))
			if err != nil {
				log.Fatalln(err)
			}
			tiles = append(tiles, img)
		}
		themeTiles[Theme(dir.Name())] = tiles
	}
}

// NewRender accepts a datastore.EventWriterCounter and returns the Renderer
// used to retrieve counts, enque write events and render PNGs.
func NewRender(d datastore.EventWriterCounter) *Renderer {
	r := &Renderer{
		datastore: d,
		save:      make(chan datastore.VisitEvent, saveQueueLength),
	}
	go r.saveLoop()
	return r
}

func (r *Renderer) saveLoop() {
	for {
		select {
		case ev := <-r.save:
			if err := r.datastore.Write(&ev); err != nil {
				log.Println("Renderer: ", err)
			}
		}
	}
}

// Count returns the amount of 'hits' for the supplied domain. Returns 0 if any
// any errors are encountered. Empty domains are not considered valid.
func (r *Renderer) Count(domain string) int {
	if domain == "" {
		return 0
	}
	q := &datastore.QueryEvent{Domain: domain}
	c, err := r.datastore.Count(q)
	if err != nil {
		log.Println("Renderer: unable to retrieve count from datstore: ", err)
		return 0
	}
	return c
}

// Add queues an event to be written to the datastore. Initial validation is
// done to make sure that a valid domain and IP address are supplied, after
// which the event will be added to the write queue (non-blocking).
func (r *Renderer) Add(ip net.IP, domain string) error {
	if domain == "" {
		return errors.New("empty domain supplied")
	}
	if ip == nil {
		return errors.New("empty IP address supplied")
	}
	ev := &datastore.VisitEvent{
		Time:   time.Now().UTC(),
		Domain: domain,
		IP:     ip,
	}
	r.save <- *ev
	return nil
}

// Render will render the supplied number and return a PNG that represents it.
// Options can be optionally supplied, if nil, defaults will be used.
func (r *Renderer) Render(o *Options, number int) (image.Image, error) {
	if o.Width < 0 {
		o.Width = DefaultWidth
	}
	if o.Width > maxWidth {
		o.Width = maxWidth
	}

	var images []image.Image
	for _, r := range fmt.Sprintf("%0*d", o.Width, number) {
		tile := themeTiles[o.Theme][int(r-'0')]
		if tile == nil {
			return nil, fmt.Errorf("no tile found for %s:%s", o.Theme, r)
		}
		images = append(images, tile)
	}
	imageBoundX := images[0].Bounds().Dx()
	imageBoundY := images[0].Bounds().Dy()

	canvasBoundX := len(images) * imageBoundX
	canvasBoundY := imageBoundY

	canvasMaxPoint := image.Point{canvasBoundX, canvasBoundY}
	canvas := image.NewRGBA(image.Rectangle{image.Point{0, 0}, canvasMaxPoint})

	for i := range images {
		x := i % len(images)
		y := i / len(images)
		minPoint := image.Point{x * imageBoundX, y * imageBoundY}
		maxPoint := minPoint.Add(image.Point{imageBoundX, imageBoundY})
		nextGridRect := image.Rectangle{minPoint, maxPoint}
		draw.Draw(canvas, nextGridRect, images[i], image.Point{}, draw.Src)
	}
	return canvas, nil
}
