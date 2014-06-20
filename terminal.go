package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"time"
)

func update_screen(pipe_open bool, stats Stats) {
	termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	draw_header()

	y := 1
	_, height := termbox.Size()
	for _, stat := range stats {
		draw_element(y, stat)
		y += 1
		if y == height-1 {
			break
		}
	}
	draw_footer(pipe_open)
	termbox.Flush()
}

func draw_line(y int, content string) {
	length, _ := termbox.Size()
	x := 0
	for _, r := range content {
		termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorBlack)
		x += 1
		if x == length-1 {
			break
		}
	}
	return

}

func draw_element(y int, stat Stat) {
	content := fmt.Sprintf("%10d %10.2f %10.2f %10.2f %10.2f %s %s", stat.seen, stat.sum, stat.average, stat.min, stat.max, stat.last_seen.Format(time.Kitchen), stat.element)
	draw_line(y, content)
	return
}

func draw_header() {
	content := fmt.Sprintf("%10s %10s %10s %10s %10s %10s %s", "seen", "sum", "average", "min","max","last seen", "element")
	draw_line(0, content)
}

func draw_footer(pipe_open bool) {
	_, height := termbox.Size()
	content := "reading from pipe"
	if !pipe_open {
	    content = "pipe is closed"
	}
	draw_line(height-1, content)
}

func read_key(c chan termbox.Event) {
	for {
		c <- termbox.PollEvent()
	}
	return
}
