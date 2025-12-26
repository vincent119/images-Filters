package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
)

func main() {
	// 建立一個 800x600 的測試圖片
	rect := image.Rect(0, 0, 800, 600)
	img := image.NewRGBA(rect)

	// 填滿背景色 (藍色)
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(img, rect, &image.Uniform{blue}, image.Point{}, draw.Src)

	// 畫一個紅色方塊
	redRect := image.Rect(100, 100, 300, 300)
	red := color.RGBA{255, 0, 0, 255}
	draw.Draw(img, redRect, &image.Uniform{red}, image.Point{}, draw.Src)

	// 存檔
	f, err := os.Create("test.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		panic(err)
	}
	println("test.jpg generated")
}
