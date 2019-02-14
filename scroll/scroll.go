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

type ScrollRunner interface {
	Run()
}

func NewScroller(sr ScrollRunner) Scroller {
	return Scroller{
		ScrollRunner: sr,
	}
}

type Scroller struct {
	ScrollRunner
}

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
			s.ScrollRunner.Run()
		case <-kb:
			paused = !paused
		}
	}
}

func NewTextScroller(r io.Reader) TextScroller {
	return TextScroller{bufio.NewScanner(r)}
}

type TextScroller struct {
	scanner *bufio.Scanner
}

func (s TextScroller) Run() {
	if more := s.scanner.Scan(); more {
		fmt.Println(s.scanner.Text())
	}
	if err := s.scanner.Err(); err != nil {
		fmt.Printf("error during scan: %v\n", err)
		os.Exit(1)
	}
}

func NewHexScroller(r io.Reader, width int) HexScroller {
	return HexScroller{
		reader: bufio.NewReader(r),
		buf:    make([]byte, width*8),
	}
}

type HexScroller struct {
	reader *bufio.Reader
	buf    []byte
}

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
