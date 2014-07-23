package stdout

import (
	"bytes"
	"fmt"
	"github.com/mdom/topstat/stat"
	"time"
)

type Terminal struct {
	PipeOpen       bool
	Metrics        []string
	StartTime      time.Time
	StatMap        *stat.StatMap
	UpdateInterval time.Duration
}

func (t *Terminal) Run(quit chan bool) {
	tick := time.Tick(t.UpdateInterval)
	for {
		select {
		case <-tick:
			t.UpdateScreen()
		}
	}
	return
}

func (t *Terminal) UpdateScreen() {

	stats := t.StatMap.FastSort()

	y := 10
	for _, stat := range stats {
		t.drawElement(stat)
		if y == 0  {
			break
		}
		y--
	}
}

func (t *Terminal) drawElement(stat stat.Stat) {

	var line myBuffer
	for _, metric := range t.Metrics {
		switch metric {
		case "sum":
			line.WriteFormat("%.2f", stat.Sum)
		case "percentage":
			line.WriteFormat("%.2f", stat.GetPercentage())
		case "rate":
			line.WriteFormat("%.2f", stat.GetRate(t.StartTime))
		case "average":
			line.WriteFormat("%.2f", stat.Average)
		case "seen":
			line.WriteFormat("%d", stat.Seen)
		case "min":
			line.WriteFormat("%.2f", stat.Min)
		case "max":
			line.WriteFormat("%.2f", stat.Max)
		case "decay":
			line.WriteFormat("%.2f", stat.Decay)
		case "lastseen":
			line.WriteFormat("%s", stat.LastSeen.Format(time.Kitchen))
		}
		line.WriteString(" ")
	}
	line.WriteFormat("%s", stat.Element)
	fmt.Println(line.String())
	return
}

func (t *Terminal) SetPipeOpen(state bool) {
	t.PipeOpen = state
	return
}


type myBuffer struct {
	bytes.Buffer
}

func (b *myBuffer) WriteFormat(format string, thing interface{}) {
	b.WriteString(fmt.Sprintf(format, thing))
}
