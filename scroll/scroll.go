package scroll

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tb "github.com/nsf/termbox-go"
)

// Runner describes an interface consisting of a single Run method.
// Implementations of Runner are expected to process a single "line" of
// input for each call to Run. At the moment, it is acceptable to exit from
// within Run, though this is likely to change.
type Runner interface {
	Run()
}

// NewScroller returns a Scroller that uses the provided Runner to process input.
func NewScroller(sr Runner) Scroller {
	return Scroller{
		Runner: sr,
	}
}

// Scroller manages the event loop for processing input using a Runner. It
// listens for key events to pause or exit the program, and calls Runner on the
// user-specified schedule.
type Scroller struct {
	Runner
}

// Start runs two goroutines in parallel. The first waits for keyboard input to
// either pause or exist the program. The second runs the primary processing
// loop which calls the underlying Runner's Run method with the provided
// frequency.
func (s Scroller) Start(speed time.Duration) {
	kb := make(chan int, 1)
	if err := tb.Init(); err != nil {
		fmt.Printf("failed to initialize input handler: %v\n", err)
		os.Exit(1)
	}
	defer tb.Close()

	go func() {
		for {
			event := tb.PollEvent()
			switch {
			case event.Key == tb.KeySpace:
				kb <- 1
			case event.Key == tb.KeyEsc:
				os.Exit(1)
			}
		}
	}()

	var paused bool
	tick := time.NewTicker(speed)
	for {
		select {
		case <-tick.C:
			if paused {
				continue
			}
			s.Runner.Run()
		case <-kb:
			paused = !paused
		}
	}
}

// NewTextScroller returns an implementation of Runner that simply outputs the
// lines of the input reader as text.
func NewTextScroller(r io.Reader) TextScroller {
	return TextScroller{bufio.NewScanner(r)}
}

// TextScroller implements Runner and simply outputs the lines of the input
// reader as text.
type TextScroller struct {
	scanner *bufio.Scanner
}

// Run implements Runner and simply scans successive lines of text and outputs
// them unmodified to stdout.
func (s TextScroller) Run() {
	if more := s.scanner.Scan(); more {
		fmt.Println(s.scanner.Text())
	}
	if err := s.scanner.Err(); err != nil {
		fmt.Printf("error during scan: %v\n", err)
		os.Exit(1)
	}
}

// NewHexScroller returns an implementation of Runner that outputs the hex
// representation of input from r in w columns of 8 bytes each, along with the
// string representation similar to the output of hexdump.
func NewHexScroller(r io.Reader, w int) HexScroller {
	return HexScroller{
		reader: bufio.NewReader(r),
		buf:    make([]byte, w*8),
	}
}

// HexScroller implements Runner and outputs the hex representation of input
// from r in w columns of 8 bytes each, along with the string representation
// similar to the output of hexdump.
type HexScroller struct {
	reader *bufio.Reader
	buf    []byte
}

// Run implements Runner and reads the bytes of the input in 8*width sized
// chunks, displaying them cleanly alongside their string representation.
func (s HexScroller) Run() {
	if _, err := s.reader.Read(s.buf); err != nil {
		if err == io.EOF {
			os.Exit(0)
		}
		fmt.Printf("error during read: %v\n", err)
		os.Exit(1)
	}
	cols := split(s.buf)
	for _, col := range cols {
		fmt.Printf("%s  ", col)
	}
	fmt.Printf("\t\t%s\n", strings.Map(replaceNPCs, string(s.buf)))
}

func replaceNPCs(r rune) rune {
	switch r {
	case '\n', '\r', ' ', '\t':
		return '.'
	default:
		return r
	}
}

type byteGroup []byte

func (g byteGroup) String() string {
	chars := make([]string, len(g))
	for i, c := range g {
		chars[i] = fmt.Sprintf("%2x", c)
	}
	return strings.Join(chars, " ")
}

func split(b []byte) []byteGroup {
	var splits []byteGroup
	for i := 0; i < len(b); i += 8 {
		splits = append(splits, b[i:i+8])
	}
	return splits
}
