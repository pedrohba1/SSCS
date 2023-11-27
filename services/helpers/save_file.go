package helpers

import (
	"image"
	"image/jpeg"
	"os"
	"strconv"
	"time"
)

func SaveToFile(img image.Image, folder string) error {
	// create file
	fname := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) + ".jpg"
	f, err := os.Create(folder + fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// convert to jpeg
	return jpeg.Encode(f, img, &jpeg.Options{
		Quality: 60,
	})
}
