package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"bytes"

	"github.com/rbrick/krieger/subtitles"
)

func main() {
	d, err := ioutil.ReadFile("Subtitle.srt")
	if err != nil {
		log.Fatal(err)
	}
	r := subtitles.NewSubripReader(bytes.NewReader(d))

	l, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(l.Text)
}
