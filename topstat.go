package main

import (
	"bufio"
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/mdom/topstat/stat"
	"github.com/mdom/topstat/view"
	"github.com/mdom/topstat/view/stdout"
	"github.com/mdom/topstat/view/termbox"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Options struct {
	Metrics      []string `short:"m" long:"metric" description:"Metrics to display" default:"sum" default:"average" value-name:"METRIC"`
	Interval     int      `short:"i" long:"interval" description:"delay between screen updates" default:"2" value-name:"INTERVAL"`
	Purge        string   `short:"p" long:"purge" description:"purge strategy" default:"decay" value-name:"STRATEGY"`
	Keep         int      `short:"k" long:"keep" description:"keep NUM elements" default:"1000" value-name:"NUM"`
	OnlyElement  bool     `short:"E" long:"only-element" description:"first element of stdin is not a number" default:"false"`
	StrictParser bool     `short:"S" long:"strict" description:"" default:"false"`
	RateUnit     string   `short:"R" long:"rate-unit" description:"per unit time" default:"minute"`
	SortOrder    string   `short:"O" long:"sort-order" description:"metric to sort by (first metric)"`
	Stdout       bool     `short:"1" long:"stdout" description:"write stats to stdout" default:"false"`
	StdoutOnce   bool     `long:"stdout-once" description:"just print stats once after pipe closed" default:"false"`
}

type Viewer interface {
	Run(chan int)
	SetPipeOpen(bool)
}

func main() {

	log.SetPrefix("topstat: ")
	log.SetFlags(0)

	var opts Options
	if _, err := flags.NewParser(&opts, flags.HelpFlag).Parse(); err != nil {
		log.Fatalln(err)
	}

	if terminal.IsTerminal(syscall.Stdin) {
		log.Fatalln("stdin can't be connected to a terminal")
	}

	sortOrder := opts.Metrics[0]
	if opts.SortOrder != "" {
		sortOrder = opts.SortOrder
	}

	statmap := &stat.StatMap{
		Stats:       make(map[string]stat.Stat),
		SortOrder:   sortOrder,
		PurgeMethod: opts.Purge,
		MaxLen:      opts.Keep,
		Tier:        10,
		Dirty:       make(map[string]bool),
		RateUnit:    opts.RateUnit,
	}

	go statmap.Decay()

	var t Viewer

	if opts.Stdout {
		t = &stdout.Viewer{
			PipeOpen:  true,
			Metrics:   opts.Metrics,
			StartTime: time.Now(),
			StatMap:   statmap,
			Once:      opts.StdoutOnce,
		}
	} else {
		t = &termbox.Viewer{
			PipeOpen:  true,
			Metrics:   opts.Metrics,
			StartTime: time.Now(),
			StatMap:   statmap,
		}
	}

	event := make(chan int)
	go t.Run(event)

	newLine := make(chan string)
	go readLine(bufio.NewReader(os.Stdin), newLine)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	tick := time.Tick(time.Duration(opts.Interval) * time.Second)

loop:
	for {
		select {
		case eventType := <-event:
			switch eventType {
			case view.Quit:
				break loop
			}
		case <-interrupt:
			event <- view.Interrupt
		case <-tick:
			event <- view.Tick
			statmap.Purge()
		case line, lineOk := <-newLine:
			if lineOk {
				var num float64
				var element string
				var err error
				if opts.OnlyElement {
					num = 0
					element = line
				} else {
					num, element, err = splitLine(line)
					if err != nil {
						if opts.StrictParser {
							// termbox.Close()
							log.Fatalln(err)
						}
					}
				}
				statmap.UpdateElement(num, element)
			} else {
				newLine = nil
				event <- view.PipeClosed
			}
		}
	}
}

func readLine(reader *bufio.Reader, c chan string) (num float64, element string, err error) {

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			break
		}
		c <- line
	}
	close(c)
	return
}

func splitLine(line string) (num float64, element string, err error) {
	line = strings.Trim(line, " \n")
	z := regexp.MustCompile(" +").Split(line, 2)

	if len(z) != 2 {
		err = errors.New(fmt.Sprintf("Cannot split string into two element:<%+v> len: %d\n", z, len(z)))
		element = line
		return
	}

	num, err = strconv.ParseFloat(z[0], 64)
	if err != nil {
		element = line
		return
	}
	element = z[1]
	return
}
