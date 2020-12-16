package visitorcounter

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"log"
	"net"
	"time"

	"github.com/michaelmcallister/visitorcounter/datastore"
	"github.com/michaelmcallister/visitorcounter/visitorcounter/theme/aomm"
	"github.com/michaelmcallister/visitorcounter/visitorcounter/theme/segment"
)

const (
	DefaultWidth = 5
	maxWidth     = 100
)

type Theme int

const (
	Segment Theme = iota
	Aomm
)

type Renderer struct {
	datastore datastore.EventWriterCounter
}

type Options struct {
	Theme Theme
	Width int
}

func NewRender(datastore datastore.EventWriterCounter) *Renderer {
	return &Renderer{datastore}
}

func (r *Renderer) Count(ctx context.Context, domain string) int {
	if domain == "" {
		log.Println("Empty domain supplied, returning 0 count.")
		return 0
	}
	q := &datastore.QueryEvent{Domain: domain}
	c, err := r.datastore.Count(ctx, q)
	if err != nil {
		log.Printf("Unable to retrieve count from datstore: %v", err)
		return 0
	}
	return c
}

func (r *Renderer) Add(ctx context.Context, ip net.IP, domain string) error {
	if domain == "" {
		log.Println("Empty domain supplied, refusing to add to datastore.")
		return errors.New("empty domain supplied")
	}
	if ip == nil {
		log.Println("Empty IP address supplied, refusing to add to datastore.")
		return errors.New("empty IP address supplied")
	}
	ev := &datastore.VisitEvent{
		Time:   time.Now().UTC(),
		Domain: domain,
		IP:     ip,
	}
	return r.datastore.Write(ctx, ev)
}

func (r *Renderer) Render(ctx context.Context, o *Options, number int) (image.Image, error) {
	var numset []image.Image
	switch o.Theme {
	case Aomm:
		numset = aomm.Get()
	default:
		numset = segment.Get()
	}

	if o.Width < 0 {
		o.Width = DefaultWidth
	}
	if o.Width > maxWidth {
		o.Width = maxWidth
	}

	var images []image.Image
	for _, r := range fmt.Sprintf("%0*d", o.Width, number) {
		images = append(images, numset[int(r-'0')])
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
