package visualizer

import (
	"image"
	"image/jpeg"
	"os"
	"strconv"
	"time"
)

func saveToFile(img image.Image) error {
	// create file
	fname := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) + ".jpg"
	f, err := os.Create("./thumbs/" + fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// convert to jpeg
	return jpeg.Encode(f, img, &jpeg.Options{
		Quality: 60,
	})
}
