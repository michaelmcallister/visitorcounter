package segment

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
)

var numbers []image.Image

func init() {
	numbers = make([]image.Image, 10)
	for i := 0; i < 10; i++ {
		f, err := os.Open(fmt.Sprintf("visitorcounter/theme/segment/%d.png", i))
		if err != nil {
			log.Fatal(err)
			return
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			log.Fatal(err)
			return
		}
		numbers[i] = img
	}
}

func Get() []image.Image {
	return numbers
}
