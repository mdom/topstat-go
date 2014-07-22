package main

import (
	"bufio"
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/mdom/topstat/stat"
	tui "github.com/mdom/topstat/terminal"
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

	t := tui.Terminal{
		PipeOpen:       true,
		Metrics:        opts.Metrics,
		StartTime:      time.Now(),
		StatMap:        statmap,
		UpdateInterval: time.Duration(opts.Interval) * time.Second,
	}

	newLine := make(chan string)
	tick := time.Tick(time.Duration(opts.Interval) * time.Second)

	go readLine(bufio.NewReader(os.Stdin), newLine)

	quit := make(chan bool)
	go t.Run(quit)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

loop:
	for {
		select {
		case <-quit:
			break loop
		case <-interrupt:
			break loop
		case <-tick:
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
				t.PipeOpen = false
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
