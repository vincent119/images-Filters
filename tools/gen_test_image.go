//go:build ignore
// +build ignore

package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

func main() {
	// 建立 800x600 的藍色圖片
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	blue := color.RGBA{0, 100, 200, 255}

	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, blue)
		}
	}

	// 儲存為 JPEG
	f, err := os.Create("data/images/test.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 85}); err != nil {
		panic(err)
	}

	println("建立測試圖片: data/images/test.jpg")
}
