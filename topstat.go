package main

import (
	"bufio"
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/nsf/termbox-go"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Options struct {
	Metrics  []string `short:"m" long:"metric" description:"Metrics to display" default:"sum" default:"average" value-name:"METRIC"`
	Interval int      `short:"i" long:"interval" description:"delay between screen updates" default:"2" value-name:"INTERVAL"`
	Purge    string   `short:"p" long:"purge" description:"purge strategy" default:"decay" value-name:"STRATEGY"`
	Keep     int      `short:"k" long:"keep" description:"keep NUM elements" default:"1000" value-name:"NUM"`
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
	pipeOpen := true

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	_, y := termbox.Size()

	statmap := &StatMap{
		stats:       make(map[string]Stat),
		sortOrder:   "sum",
		purgeMethod: opts.Purge,
		maxLen:      opts.Keep,
		tier:        y - 2,
		dirty:       make(map[string]bool),
	}

	go statmap.decay()

	newLine := make(chan string)
	keyPressed := make(chan termbox.Event)
	tick := time.Tick(time.Duration(opts.Interval) * time.Second)

	go readLine(bufio.NewReader(os.Stdin), newLine)
	go readKey(keyPressed)

loop:
	for {
		select {
		case <-tick:
			statmap.purge()
			updateScreen(pipeOpen, opts.Metrics, statmap.fastsort())
		case event := <-keyPressed:
			switch event.Type {
			case termbox.EventKey:
				switch event.Ch {
				case 'q':
					break loop
				case 'a':
					statmap.SetSortOrder("average")
				case 'd':
					statmap.SetSortOrder("decay")
				case 's':
					statmap.SetSortOrder("sum")
				case 'n':
					statmap.SetSortOrder("seen")
				case '<':
					statmap.SetSortOrder("min")
				case '>':
					statmap.SetSortOrder("max")
				case 'l':
					statmap.SetSortOrder("last_seen")
				}
				switch event.Ch {
				case 'l', 'a', 'd', 's', 'n', '<', '>':
					updateScreen(pipeOpen, opts.Metrics, statmap.fastsort())
				}
			case termbox.EventResize:
				_, y := termbox.Size()
				statmap.setTier(y - 2)
				updateScreen(pipeOpen, opts.Metrics, statmap.fastsort())
			}
		case line, lineOk := <-newLine:
			if lineOk {
				num, element, err := splitLine(line)
				if err == nil {
					statmap.updateElement(num, element)
				}
			} else {
				newLine = nil
				pipeOpen = false
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
		return
	}

	num, err = strconv.ParseFloat(z[0], 64)
	if err != nil {
		return
	}
	element = z[1]
	return
}
