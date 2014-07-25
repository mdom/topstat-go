package stdout

import (
	"bytes"
	"fmt"
	"github.com/mdom/topstat/stat"
	"github.com/mdom/topstat/view"
	"time"
)

type Viewer struct {
	PipeOpen       bool
	Metrics        []string
	StartTime      time.Time
	StatMap        *stat.StatMap
	UpdateInterval time.Duration
	needNewline    bool
	Once           bool
}

func (t *Viewer) Run(event chan int) {
	tick := time.Tick(t.UpdateInterval)
	for {
		select {
		case eventType := <-event:
			switch eventType {
			case view.Interrupt:
				t.Once = false
				if t.needNewline {
					t.UpdateScreen()
				}
			case view.PipeClosed:
				t.Once = false
				t.UpdateScreen()
				event <- view.Quit
			}
		case <-tick:
			if t.PipeOpen {
				t.UpdateScreen()
			}
		}
	}
	return
}

func (t *Viewer) UpdateScreen() {

	if t.Once {
		return
	}

	if t.needNewline {
		fmt.Println()
	} else {
		t.needNewline = true
		
	}

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

func (t *Viewer) drawElement(stat stat.Stat) {

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

func (t *Viewer) SetPipeOpen(state bool) {
	t.PipeOpen = state
	return
}


type myBuffer struct {
	bytes.Buffer
}

func (b *myBuffer) WriteFormat(format string, thing interface{}) {
	b.WriteString(fmt.Sprintf(format, thing))
}
