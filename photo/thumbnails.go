package photo

import (
	"github.com/nfnt/resize"
	"image/jpeg"
	"log"
	"os"
)

type ThumbnailManager interface {
	Execute(src, dst string) error
}

type ThumbnailEngine struct {
	Width  uint
	Height uint
}

// Make thumbnail from src image.
func (t *ThumbnailEngine) Execute(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	m := resize.Resize(t.Width, t.Height, img, resize.Lanczos3)

	out, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	return jpeg.Encode(out, m, nil)
}
