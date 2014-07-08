package main

import "sort"
import "time"
import "sync"

type Stat struct {
	sum      float64
	average  float64
	min      float64
	max      float64
	lastSeen time.Time
	seen     int
	element  string
	decay    float64
}

type StatMap struct {
	sync.Mutex
	stats       map[string]Stat
	sortOrder   string
	purgeMethod string
	maxLen      int
	tier        int
	forceResort bool
	dirty       map[string]bool
}

type Stats []Stat

type BySum []Stat
type ByAverage []Stat
type BySeen []Stat
type ByMax []Stat
type ByMin []Stat
type ByLastSeen []Stat
type ByDecay []Stat

func (s BySum) Len() int           { return len(s) }
func (s BySum) Less(i, j int) bool { return s[i].sum > s[j].sum }
func (s BySum) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByAverage) Len() int           { return len(s) }
func (s ByAverage) Less(i, j int) bool { return s[i].average > s[j].average }
func (s ByAverage) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s BySeen) Len() int           { return len(s) }
func (s BySeen) Less(i, j int) bool { return s[i].seen > s[j].seen }
func (s BySeen) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByMax) Len() int           { return len(s) }
func (s ByMax) Less(i, j int) bool { return s[i].max > s[j].max }
func (s ByMax) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByMin) Len() int           { return len(s) }
func (s ByMin) Less(i, j int) bool { return s[i].min < s[j].min }
func (s ByMin) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByLastSeen) Len() int           { return len(s) }
func (s ByLastSeen) Less(i, j int) bool { return s[i].lastSeen.After(s[j].lastSeen) }
func (s ByLastSeen) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByDecay) Len() int           { return len(s) }
func (s ByDecay) Less(i, j int) bool { return s[i].decay > s[j].decay }
func (s ByDecay) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (statmap *StatMap) decay() {
	c := time.Tick(1 * time.Second)
	n := 0
	for _ = range c {
		if n < 60 {
			n++
		}

		statmap.Lock()
		for m, stat := range statmap.stats {
			stat.decay = (1.0/3.0 - stat.decay) / float64(n)
			statmap.stats[m] = stat

		}
		statmap.Unlock()
	}
	return
}

func (statmap *StatMap) SetSortOrder(sortOrder string) {
	statmap.sortOrder = sortOrder
	statmap.forceResort = true
	return
}

func (statmap *StatMap) SetTier(tier int) {
	statmap.tier = tier
	statmap.forceResort = true
	return
}

func (statmap *StatMap) sort() Stats {

	statmap.Lock()
	defer statmap.Unlock()

	var s Stats
	for _, stat := range statmap.stats {
		s = append(s, stat)
	}
	s.sort(statmap.sortOrder)
	return s
}

func (s Stats) sort(sortOrder string) {
	switch sortOrder {
	case "sum":
		sort.Stable(BySum(s))
	case "average":
		sort.Stable(ByAverage(s))
	case "decay":
		sort.Stable(ByDecay(s))
	case "rate":
		sort.Stable(BySeen(s))
	case "seen":
		sort.Stable(BySeen(s))
	case "max":
		sort.Stable(ByMax(s))
	case "min":
		sort.Stable(ByMin(s))
	case "last_seen":
		sort.Stable(ByLastSeen(s))
	}
	return
}

func (statmap *StatMap) purge() (purged bool) {
	if statmap.maxLen == -1 {
		return false
	}

	var s Stats
	for _, stat := range statmap.stats {
		s = append(s, stat)
	}
	s.sort(statmap.purgeMethod)

	statmap.Lock()
	defer statmap.Unlock()

	if len(s) > statmap.maxLen {
		s = s[statmap.maxLen:len(s)]
		for _, stat := range s {
			delete(statmap.stats, stat.element)
		}
		return true
	}

	return false
}

func (statmap *StatMap) fastsort() Stats {

	n := statmap.tier

	var s Stats

	if statmap.forceResort {
		statmap.dirty = make(map[string]bool)
		s = statmap.sort()
	} else {
		statmap.Lock()
		for element := range statmap.dirty {
			s = append(s, statmap.stats[element])
		}
		statmap.Unlock()
		s.sort(statmap.sortOrder)
	}

	if len(s) > n {
		s = s[:n]
	}

	statmap.dirty = make(map[string]bool)
	for _, stat := range s {
		statmap.dirty[stat.element] = true
	}

	return s
}

func (statmap *StatMap) updateElement(num float64, element string) (err error) {

	statmap.Lock()
	defer statmap.Unlock()

	stat, ok := statmap.stats[element]

	if !ok {
		stat = Stat{}
	}

	max := stat.max
	min := stat.min
	if num > max {
		max = num
	}
	if num < min {
		min = num
	}

	statmap.stats[element] = Stat{
		sum:      stat.sum + num,
		average:  ((stat.average*float64(stat.seen) + num) / (float64(stat.seen) + 1)),
		seen:     stat.seen + 1,
		element:  element,
		min:      min,
		max:      max,
		lastSeen: time.Now(),
		decay:    stat.decay + 1,
	}
	statmap.dirty[element] = true
	return
}
