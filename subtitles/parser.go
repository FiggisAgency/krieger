package subtitles

import (
	"io"

	"bufio"

	"github.com/rbrick/krieger/subtitles/subrip"
)

//NewSubripReader creates a new SubRip parser
func NewSubripReader(r io.Reader) *subrip.Reader {
	return &subrip.Reader{
		Buf: bufio.NewReader(r),
	}
}
