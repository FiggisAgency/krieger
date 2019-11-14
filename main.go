package main

import (
	"fmt"
	"github.com/FiggisAgency/krieger/av"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func FastCompare(img1, img2 *image.RGBA) (int64, error) {
	if img1.Bounds() != img2.Bounds() {
		return 0, fmt.Errorf("image bounds not equal: %+v, %+v", img1.Bounds(), img2.Bounds())
	}

	accumError := int64(0)

	for i := 0; i < len(img1.Pix); i++ {
		accumError += int64(sqDiffUInt8(img1.Pix[i], img2.Pix[i]))
	}

	return int64(math.Sqrt(float64(accumError))), nil
}

func sqDiffUInt8(x, y uint8) uint64 {
	d := uint64(x) - uint64(y)
	return d * d
}

type averageRGB struct {
	R, G, B int64
}

func (a *averageRGB) RGBA(pixelCount int64) color.RGBA {
	return color.RGBA{
		R: uint8(a.R / pixelCount),
		G: uint8(a.G / pixelCount),
		B: uint8(a.B / pixelCount),
		A: uint8(255),
	}
}
func Compare(img1, img2 *image.RGBA, frame int64) int64 {
	var averageDiff int64

	fmt.Println(img1.Bounds())
	fmt.Println(img1.Bounds().Dx() / 100)

	//subImgCount := 0
	for w := 0; w < img1.Bounds().Dx(); w += 100 {
		for h := 0; h < img1.Bounds().Dy(); h += 100 {

			argba := averageRGB{}
			pixelCount := int64(0)
			for x := w; x <= w+100; x++ {
				for y := h; y <= h+100; y++ {
					rgba := img1.RGBAAt(x, y)
					argba.R += int64(rgba.R)
					argba.G += int64(rgba.G)
					argba.B += int64(rgba.B)
					pixelCount++
				}
			}

			for x := w; x <= w+100; x++ {
				for y := h; y <= h+100; y++ {
					img1.SetRGBA(x, y, argba.RGBA(pixelCount))
					img1.SetRGBA(x, h, color.RGBA{0, 0, 0, 255})
					img1.SetRGBA(w, y, color.RGBA{0, 0, 0, 255})
				}
			}

		}
	}

	f, _ := os.Create(fmt.Sprintf("work/grid_%d.png", frame))
	png.Encode(f, img1)
	return averageDiff
}

func main() {
	video := av.LoadVideo("max.mp4")
	defer video.Cleanup()

	frameCount := int64(0)
	//var total int64
	var lastFrame *image.RGBA

	initialTarget := int64(50)

	for frame, err := video.ReadFrame(); err == nil; {
		if lastFrame == nil {
			lastFrame = frame

			//f, _ := os.Create(fmt.Sprintf("work/test_%d.png", frameCount))
			//
			//_ = png.Encode(f, frame)
		} else {
			if frameCount >= initialTarget {
				break
			}

			//diff, err := FastCompare(frame, lastFrame)

			Compare(lastFrame, frame, frameCount)

			frameCount++

			//fmt.Println("diff between frames", frameCount, "and", frameCount-1, "is", diff)

			//f, _ := os.Create(fmt.Sprintf("work/test_%d.png", frameCount))
			//
			//_ = png.Encode(f, frame)

			lastFrame = frame
			frame, err = video.ReadFrame()

			//break
			if frame == nil {
				break
			}

		}

	}

	//fmt.Println("average difference between frames:", total/frameCount)
	//fmt.Println("saved", frameCount, "/1000")
}
