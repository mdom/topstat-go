package main

import (
	"bytes"
	"fmt"
	"github.com/nsf/termbox-go"
	"time"
)

func updateScreen(pipeOpen bool, metrics []string, stats Stats) {

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawHeader(metrics)

	y := 1
	_, height := termbox.Size()
	for _, stat := range stats {
		drawElement(y, stat, metrics)
		y += 1
		if y == height-1 {
			break
		}
	}
	drawFooter(pipeOpen)
	termbox.Flush()
}

func drawLine(y int, content string, bg termbox.Attribute) {
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

func drawElement(y int, stat Stat, metrics []string) {

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
		case "decay":
			line.WriteFormat("%10.2f", stat.decay)
		case "lastseen":
			line.WriteFormat("%10s", stat.lastSeen.Format(time.Kitchen))
		}
		line.WriteString(" ")
	}
	line.WriteFormat("%s", stat.element)
	drawLine(y, line.String(), termbox.ColorDefault)
	return
}

type MyBuffer struct {
	bytes.Buffer
}

func (b *MyBuffer) WriteFormat(format string, thing interface{}) {
	b.WriteString(fmt.Sprintf(format, thing))
}

func drawHeader(metrics []string) {
	var line MyBuffer
	for _, metric := range metrics {
		line.WriteFormat("%10s ", metric)
	}
	line.WriteString("element")
	drawLine(0, line.String(), termbox.ColorDefault|termbox.AttrReverse)
}

func drawFooter(pipeOpen bool) {
	_, height := termbox.Size()
	content := "reading from pipe"
	if !pipeOpen {
		content = "pipe is closed"
	}
	drawLine(height-1, content, termbox.ColorDefault|termbox.AttrReverse)
}

func readKey(c chan termbox.Event) {
	for {
		c <- termbox.PollEvent()
	}
	return
}
