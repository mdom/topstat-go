package main

import (
	"bytes"
	"fmt"
	"github.com/nsf/termbox-go"
	"time"
)

type Terminal struct {
	pipeOpen  *bool
	metrics   []string
	startTime time.Time
}

func (t *Terminal) updateScreen(stats Stats) {

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	t.drawHeader()

	y := 1
	_, height := termbox.Size()
	for _, stat := range stats {
		t.drawElement(y, stat)
		y += 1
		if y == height-1 {
			break
		}
	}
	t.drawFooter(len(stats))
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

func (t *Terminal) drawElement(y int, stat Stat) {

	var line MyBuffer
	for _, metric := range t.metrics {
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

func (t *Terminal) drawHeader() {
	var line MyBuffer
	for _, metric := range t.metrics {
		line.WriteFormat("%10s ", metric)
	}
	line.WriteString("element")
	drawLine(0, line.String(), termbox.ColorDefault|termbox.AttrReverse)
}

func (t *Terminal) drawFooter(len int) {
	_, height := termbox.Size()
	pipeState := "open"
	if *t.pipeOpen == false {
		pipeState = "closed"
	}

	content := fmt.Sprintf(
		"Pipe: %s | Elapsed: %s | Elements: %d",
		pipeState,
		time.Since(t.startTime).String(),
		len,
	)

	drawLine(height-1, content, termbox.ColorDefault|termbox.AttrReverse)
}

func readKey(c chan termbox.Event) {
	for {
		c <- termbox.PollEvent()
	}
	return
}
