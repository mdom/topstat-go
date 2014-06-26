package main

import (
	"bytes"
	"fmt"
	"github.com/nsf/termbox-go"
	"time"
)

func update_screen(pipe_open bool, metrics []string, stats Stats) {

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	draw_header(metrics)

	y := 1
	_, height := termbox.Size()
	for _, stat := range stats {
		draw_element(y, stat, metrics)
		y += 1
		if y == height-1 {
			break
		}
	}
	draw_footer(pipe_open)
	termbox.Flush()
}

func draw_line(y int, content string, bg termbox.Attribute) {
	length, _ := termbox.Size()
	x := 0
	for _, r := range content {
		termbox.SetCell(x, y, r, termbox.ColorDefault, bg)
		x += 1
		if x == length-1 {
			break
		}
	}
	for ; x < length; x++ {
		termbox.SetCell(x, y, ' ', termbox.ColorDefault, bg)
	}
	return

}

func draw_element(y int, stat Stat, metrics []string) {

	var line MyBuffer
	for _, metric := range metrics {
		switch metric {
		case "sum":
			line.WriteFormat("%10.2f", stat.sum)
		case "average":
			line.WriteFormat("%10.2f", stat.average)
		case "seen":
			line.WriteFormat("%10d", stat.seen)
		case "min":
			line.WriteFormat("%10.2f", stat.min)
		case "max":
			line.WriteFormat("%10.2f", stat.max)
		case "lastseen":
			line.WriteFormat("%10s", stat.last_seen.Format(time.Kitchen))
		}
		line.WriteString(" ")
	}
	line.WriteFormat("%s", stat.element)
	draw_line(y, line.String(), termbox.ColorDefault)
	return
}

type MyBuffer struct {
	bytes.Buffer
}

func (b *MyBuffer) WriteFormat(format string, thing interface{}) {
	b.WriteString(fmt.Sprintf(format, thing))
}

func draw_header(metrics []string) {
	var line bytes.Buffer
	for _, metric := range metrics {
		line.WriteString(fmt.Sprintf("%10s ", metric))
	}
	line.WriteString("element")
	draw_line(0, line.String(), termbox.ColorDefault|termbox.AttrReverse)
}

func draw_footer(pipe_open bool) {
	_, height := termbox.Size()
	content := "reading from pipe"
	if !pipe_open {
		content = "pipe is closed"
	}
	draw_line(height-1, content, termbox.ColorDefault|termbox.AttrReverse)
}

func read_key(c chan termbox.Event) {
	for {
		c <- termbox.PollEvent()
	}
	return
}
