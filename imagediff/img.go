package imagediff

import (
	"errors"
	"image"
	"image/color"
	"math"
)

type Bucket struct {
	R, G, B    int64
	PixelCount int64
}

func (bucket *Bucket) Reset() {
	bucket.R = 0
	bucket.G = 0
	bucket.B = 0
	bucket.PixelCount = 0
}

func (bucket *Bucket) RGBA() color.RGBA {
	return color.RGBA{
		R: uint8(bucket.R / bucket.PixelCount),
		G: uint8(bucket.G / bucket.PixelCount),
		B: uint8(bucket.B / bucket.PixelCount),
		A: 0xFF,
	}
}

func (bucket *Bucket) RGB() int64 {
	rgba := bucket.RGBA()

	return int64((rgba.R) + (rgba.B) + (rgba.G))
}

func (bucket *Bucket) Add(rgba color.RGBA) {

	bucket.R += int64(rgba.R)
	bucket.G += int64(rgba.G)
	bucket.B += int64(rgba.B)

	bucket.PixelCount++
}

func NewBucket() *Bucket {
	return new(Bucket)
}

func DiffBucket(b1, b2 *Bucket) int64 {
	avr1, avr2 := b1.RGBA(), b2.RGBA()

	rd := math.Abs(float64(avr1.R - avr2.R))
	gd := math.Abs(float64(avr1.G - avr2.G))
	bd := math.Abs(float64(avr1.B - avr2.B))

	return int64(math.Sqrt((rd * rd) + (gd * gd) + (bd * bd)))
}

// the lower the diff, the more similar the images are
func Diff(b1, b2 []*Bucket) (int64, error) {
	var totalDiff int64

	if len(b1) != len(b2) {
		return 0, errors.New("lengths are not the same")
	}

	for i := 0; i < len(b1); i++ {
		totalDiff += DiffBucket(b1[i], b2[i])
	}
	return totalDiff, nil
}

func DiffImages(img1, img2 *image.RGBA) (int64, error) {
	if img1.Bounds() != img2.Bounds() {
		return 0, errors.New("images are not the same size")
	}

	bucket1, bucket2 := NewBucket(), NewBucket()
	var diff int64
	for w := 0; w < img1.Bounds().Dx(); w += 100 {
		for h := 0; h < img1.Bounds().Dy(); h += 100 {
			for x := w; x <= w+100; x++ {
				for y := h; y <= h+100; y++ {
					rgba1, rgba2 := img1.RGBAAt(x, y), img2.RGBAAt(x, y)
					bucket1.Add(rgba1)
					bucket2.Add(rgba2)

				}
			}

			diff += DiffBucket(bucket1, bucket2)

			bucket1.Reset()
			bucket2.Reset()
		}
	}

	return diff, nil
}
