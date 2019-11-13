package main

import (
	"fmt"
	"github.com/FiggisAgency/krieger/av"
	"image/png"
	"os"
)

func main() {
	video := av.LoadVideo("test.mp4")
	defer video.Cleanup()

	for i := 0; i < 30; i++ {
		f, _ := os.Create(fmt.Sprintf("test_%d.png", i))

		png.Encode(f, video.ReadFrame())

		fmt.Println("test")
	}
}
