package terminal

import (
	"bytes"
	"fmt"
	"github.com/mdom/topstat/stat"
	"github.com/nsf/termbox-go"
	"time"
)

type Terminal struct {
	PipeOpen  bool
	Paused    bool
	Metrics   []string
	StartTime time.Time
	StatMap   *stat.StatMap
}

func (t *Terminal) UpdateScreen(stats stat.Stats) {

	if t.Paused {
		return
	}

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

func (t *Terminal) drawElement(y int, stat stat.Stat) {

	var line MyBuffer
	for _, metric := range t.Metrics {
		switch metric {
		case "sum":
			line.WriteFormat("%10.2f", stat.Sum)
		case "percentage":
			line.WriteFormat("%10.2f", stat.GetPercentage())
		case "rate":
			line.WriteFormat("%10.2f", stat.GetRate(t.StartTime))
		case "average":
			line.WriteFormat("%10.2f", stat.Average)
		case "seen":
			line.WriteFormat("%10d", stat.Seen)
		case "min":
			line.WriteFormat("%10.2f", stat.Min)
		case "max":
			line.WriteFormat("%10.2f", stat.Max)
		case "decay":
			line.WriteFormat("%10.2f", stat.Decay)
		case "lastseen":
			line.WriteFormat("%10s", stat.LastSeen.Format(time.Kitchen))
		}
		line.WriteString(" ")
	}
	line.WriteFormat("%s", stat.Element)
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
	for _, metric := range t.Metrics {
		line.WriteFormat("%10s ", metric)
	}
	line.WriteString("element")
	drawLine(0, line.String(), termbox.ColorDefault|termbox.AttrReverse)
}

func (t *Terminal) drawFooter(len int) {
	_, height := termbox.Size()
	pipeState := "open"
	if t.PipeOpen == false {
		pipeState = "closed"
	}

	content := fmt.Sprintf(
		"Pipe: %s | Elapsed: %s | Elements: %d",
		pipeState,
		time.Since(t.StartTime).String(),
		len,
	)

	drawLine(height-1, content, termbox.ColorDefault|termbox.AttrReverse)
}

func ReadKey(c chan termbox.Event) {
	for {
		c <- termbox.PollEvent()
	}
	return
}
