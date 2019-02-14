package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/haleyrc/scrollr/scroll"
)

func main() {
	var (
		mult     = flag.Int("s", 1, "The speed multiplier. Default is some lines per minute.")
		h        = flag.Bool("h", false, "Output hex instead of text.")
		w        = flag.Int("w", 2, "The width to display at once in hex output.")
		filename = flag.String("f", "", "The file to display.")
	)
	flag.Parse()

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

	speed := time.Second / time.Duration(*mult)
	if *h {
		scroll.NewScroller(scroll.NewHexScroller(f, *w)).Start(speed)
	} else {
		scroll.NewScroller(scroll.NewTextScroller(f)).Start(speed)
	}
}
