package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tb "github.com/nsf/termbox-go"
)

const DefaultRate = time.Second
const DefaultWidth = 2

func main() {
	var (
		mult     = flag.Int("s", 1, "The speed multiplier. Default is some lines per minute.")
		h        = flag.Bool("h", false, "Output hex instead of text.")
		w        = flag.Int("w", DefaultWidth, "The wdith to display at once in hex output.")
		filename = flag.String("f", "", "The file to display.")
	)
	flag.Parse()

	speed := DefaultRate / time.Duration(*mult)
	fmt.Printf("Printing %f lines / second.\n", float64(time.Second/speed))

	if *filename == "" {
		flag.Usage()
		os.Exit(2)
	}

	f, err := os.Open(*filename)
	if err != nil {
		fmt.Printf("failed to open file: %s: %v\n", *filename, err)
		os.Exit(1)
	}
	defer f.Close()

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

	tick := time.NewTicker(speed)
	var paused bool
	if *h {
		buf := bufio.NewReader(f)
		out := make([]byte, *w*8)
		for {
			select {
			case <-tick.C:
				if paused {
					continue
				}
				if _, err := buf.Read(out); err != nil {
					if err == io.EOF {
						os.Exit(0)
					}
					fmt.Printf("error during read: %s: %v\n", *filename, err)
					os.Exit(1)
				}
				cols := split(out)
				for _, col := range cols {
					fmt.Printf("%s  ", col)
				}
				fmt.Printf("\t\t%s\n", strings.Map(replaceNPCs, string(out)))
			case <-kb:
				paused = !paused
			}
		}
	} else {
		scanner := bufio.NewScanner(f)
		for {
			select {
			case <-tick.C:
				if paused {
					continue
				}
				if more := scanner.Scan(); more {
					fmt.Println(scanner.Text())
					time.Sleep(speed)
				}
				if err := scanner.Err(); err != nil {
					fmt.Printf("error during scan: %s: %v\n", *filename, err)
					os.Exit(1)
				}
			case <-kb:
				paused = !paused
			}
		}
	}
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
