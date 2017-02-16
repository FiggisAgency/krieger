package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"bytes"

	"github.com/rbrick/krieger/subtitles"
	"github.com/rbrick/krieger/subtitles/subrip"
)

func main() {
	d, err := ioutil.ReadFile("Example.srt")
	if err != nil {
		log.Fatal(err)
	}
	r := subtitles.NewSubripReader(bytes.NewReader(d))
	for i := 0; i < 10; i++ {
		l, err := r.Read()

		if err == subrip.ErrCompleted {
			break
		}

		fmt.Println(i, ":")
		fmt.Println("Appear Time:", l.FromStr)
		fmt.Println("Disappear Time:", l.ToStr)
		fmt.Println("Text:", l.Text)
		fmt.Println()
		fmt.Println()
	}
}
