package subrip

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strconv"
	"time"

	"strings"

	"github.com/rbrick/krieger/utils"
)

var (
	//ErrEmptyLine means the current line is empty
	ErrEmptyLine = errors.New("Line is empty")
	//ErrCompleted means the reader has completed
	ErrCompleted = errors.New("Completed reading")
)

// The timestamp pattern for SubRip
var tsPattern = regexp.MustCompile("(?P<hours>[0-9]{2,}):(?P<mins>[0-9]{2}):(?P<secs>[0-9]{2}),(?P<ms>[0-9]{3})")

//Reader represents a SubRip Parser
type Reader struct {
	Buf *bufio.Reader
}

//Line represents a line of dialogue
type Line struct {
	// The sequence in the scene
	Seq int64
	// The time the line appears as a string
	FromStr string
	// The time the line appears
	From int64
	// The time the line disappears as a string
	ToStr string
	// The time the line disappears
	To int64
	// The actual dialogue
	Text string
}

//Read reads a line
func (r *Reader) Read() (line *Line, err error) {
	s, lErr := r.Buf.ReadString(byte('\n'))
	s = strings.Replace(strings.TrimSpace(s), string(rune(0xFEFF)), "", -1) // replace the byte-order mark (0xFEFF)
	if s == "" {
		err = ErrEmptyLine
		return
	} else if lErr == io.EOF {
		err = ErrCompleted
		return
	}

	line = &Line{}

	for s != "" && err == nil {
		// Assume we are parsing sequence
		if i, iErr := strconv.Atoi(s); iErr == nil {
			line.Seq = int64(i)
		} else if ss, low, high := parseTimestamps(s); low != -1 && high != -1 {
			line.FromStr = ss[0]
			line.From = low
			line.ToStr = ss[1]
			line.To = high
		} else {
			line.Text += string('\n') + s
		}
		s, lErr = r.Buf.ReadString(byte('\n'))
		s = strings.TrimSpace(s)
	}

	return
}

//00:20:41,150 --> 00:20:45,109
func parseTimestamps(s string) ([]string, int64, int64) {
	ss := strings.Split(strings.Replace(s, " ", "", -1), "-->")
	if len(ss) < 2 {
		return ss, -1, -1
	}
	return ss, parseTimestamp(ss[0]), parseTimestamp(ss[1])
}

func parseTimestamp(src string) int64 {
	regexRes := utils.MatchRegex(tsPattern, src)
	h, _ := strconv.Atoi(regexRes["hours"])
	m, _ := strconv.Atoi(regexRes["mins"])
	s, _ := strconv.Atoi(regexRes["secs"])
	ms, _ := strconv.Atoi(regexRes["ms"])

	ms += int((time.Duration(h) * time.Hour).Seconds()) * 1000
	ms += int((time.Duration(m) * time.Minute).Seconds()) * 1000
	ms += s * 1000
	return int64(ms)
}
