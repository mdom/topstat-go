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
	pipe_open := true

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	_, y := termbox.Size()

	statmap := &StatMap{
		stats:        make(map[string]Stat),
		sort_order:   "sum",
		purge_method: opts.Purge,
		max_len:      opts.Keep,
		top_n:        y - 2,
	}

	go statmap.decay()

	new_line := make(chan string)
	key_pressed := make(chan termbox.Event)
	tick := time.Tick(time.Duration(opts.Interval) * time.Second)

	go read_line(bufio.NewReader(os.Stdin), new_line)
	go read_key(key_pressed)

loop:
	for {
		select {
		case <-tick:
			statmap.purge()
			update_screen(pipe_open, opts.Metrics, statmap.fastsort())
		case event := <-key_pressed:
			switch event.Type {
			case termbox.EventKey:
				switch event.Ch {
				case 'q':
					break loop
				case 'a':
					statmap.sort_order = "average"
				case 'd':
					statmap.sort_order = "decay"
				case 's':
					statmap.sort_order = "sum"
				case 'n':
					statmap.sort_order = "seen"
				case '<':
					statmap.sort_order = "min"
				case '>':
					statmap.sort_order = "max"
				case 'l':
					statmap.sort_order = "last_seen"
				}
				switch event.Ch {
				case 'l', 'a', 'd', 's', 'n', '<', '>':
					update_screen(pipe_open, opts.Metrics, statmap.fastsort())
				}
			case termbox.EventResize:
				_, y := termbox.Size()
				statmap.top_n = y - 2
				update_screen(pipe_open, opts.Metrics, statmap.fastsort())
			}
		case line, line_ok := <-new_line:
			if line_ok {
				statmap.update_element(line)
			} else {
				new_line = nil
				pipe_open = false
			}
		}
	}
}

func read_line(reader *bufio.Reader, c chan string) (num float64, element string, err error) {

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

func split_line(line string) (element string, num float64, err error) {
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
