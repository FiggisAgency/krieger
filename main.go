package main

import (
	"fmt"
	"github.com/FiggisAgency/krieger/av"
	"github.com/FiggisAgency/krieger/imagediff"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sort"
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

			//argba := averageRGB{}
			//pixelCount := int64(0)

			bucket := imagediff.NewBucket()
			for x := w; x <= w+100; x++ {
				for y := h; y <= h+100; y++ {
					rgba := img1.RGBAAt(x, y)
					bucket.Add(rgba)
				}
			}

			//fmt.Printf("argba: %v, bucket: %v\n", argba, bucket)

			for x := w; x <= w+100; x++ {
				for y := h; y <= h+100; y++ {
					img1.SetRGBA(x, y, bucket.RGBA())
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

type colorAverageArray []*imagediff.Bucket

// Len is the number of elements in the collection.
func (p colorAverageArray) Len() int {
	return len(p)
}

func (p colorAverageArray) Less(i, j int) bool {
	return p[i].RGB() < p[j].RGB()
}

// Swap swaps the elements with indexes i and j.
func (p colorAverageArray) Swap(i, j int) {
	p[j], p[i] = p[i], p[j]
}

func main() {
	video := av.LoadVideo(`/mnt/d/Movies/Harold and Maude (1971).mkv`)
	defer video.Cleanup()

	//frameCount, savedFrames := int64(0), int64(0)
	//var totalDiff int64
	//var lastSavedFrame *image.RGBA

	frame, _ := video.ReadFrame()

	var colorAverages colorAverageArray

	theScene := image.NewRGBA(frame.Bounds())

	for w := 0; w < frame.Bounds().Dx(); w += 100 {
		for h := 0; h < frame.Bounds().Dy(); h += 100 {
			bucket := imagediff.NewBucket()
			for x := w; x <= w+100; x++ {
				for y := h; y <= h+100; y++ {
					bucket.Add(frame.RGBAAt(x, y))
				}
			}

			for x := w; x <= w+100; x++ {
				for y := h; y <= h+100; y++ {
					theScene.SetRGBA(x, y, bucket.RGBA())
					theScene.SetRGBA(x, h, color.RGBA{0, 0, 0, 255})
					theScene.SetRGBA(w, y, color.RGBA{0, 0, 0, 255})
				}
			}

			colorAverages = append(colorAverages, bucket)
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, 1600, 200))

	counts := map[int64]int{}
	for _, v := range colorAverages {
		if _, ok := counts[v.RGB()]; ok {
			fmt.Println("seen rgb:", v.RGB())
			counts[v.RGB()] += 1
		} else {
			counts[v.RGB()] = 1
		}
	}

	fmt.Println(len(counts))

	keys := make([]int64, len(counts))
	for k, _ := range counts {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return counts[keys[i]] > counts[keys[j]]
	})

	fmt.Println(keys)

	drawn := 0
	for i := 0; i < len(keys); i++ {
		if keys[i] == 0 {
			fmt.Println("k == 0", i)
			continue
		}
		k := keys[i]
		fmt.Println("drawing")

		if drawn < 8 {
			for y := 0; y < 200; y++ {
				for x := 0; x < 200; x++ {

					img.Set(x+(drawn*200), y, color.RGBA{
						R: uint8((k >> 16) & 0xff),
						G: uint8((k >> 8) & 0xff),
						B: uint8(k & 0xff),
						A: 0xFF,
					})
				}
			}
			drawn++
		} else {
			break
		}

	}

	f, _ := os.Create("average_colors.png")
	_ = png.Encode(f, img)

	f, _ = os.Create("the_frame.png")
	_ = png.Encode(f, frame)

	f, _ = os.Create("the_scene.png")
	_ = png.Encode(f, theScene)

	//for frame, err := video.ReadFrame(); err == nil; {
	//	if lastSavedFrame == nil {
	//		lastSavedFrame = frame
	//
	//		f, _ := os.Create(fmt.Sprintf("work/test%03d.png", frameCount))
	//
	//		_ = png.Encode(f, frame)
	//	} else {
	//
	//		//diff, _ := FastCompare(frame, lastSavedFrame)
	//
	//		diff, _ := imagediff.DiffImages(frame, lastSavedFrame)
	//
	//		//fmt.Println("diff:", diff, "error:", err)
	//
	//		frameCount++
	//
	//		totalDiff += diff
	//		fmt.Println("diff between frames", frameCount, "and", frameCount-1, "is", diff)
	//
	//		if diff >= 45000 {
	//			savedFrames++
	//			lastSavedFrame = frame
	//			f, _ := os.Create(fmt.Sprintf("work/test%03d.png", savedFrames))
	//
	//			_ = png.Encode(f, frame)
	//		}
	//
	//		if frameCount >= 250 {
	//			break
	//		}
	//
	//		frame, err = video.ReadFrame()
	//
	//		//break
	//		if frame == nil {
	//			break
	//		}
	//
	//	}
	//
	//}

	//for i := 0; i < 10; i++ {
	//	bucket := imagediff.Bucket{}
	//	bucket.Add(color.RGBA{0, 0, 1, 0xFF})
	//	fmt.Printf("bucket addr: %p, B addr: %p, B value: %v\n", &bucket, &bucket.G, bucket.G)
	//}

	//fmt.Println("average difference between frames:", totalDiff/frameCount)
	//fmt.Println("saved", savedFrames, "/", frameCount)
}
