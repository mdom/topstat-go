package stat

import (
	"math"
	"sort"
	"sync"
	"time"
)

type Stat struct {
	Sum      float64
	Average  float64
	Min      float64
	Max      float64
	LastSeen time.Time
	Seen     int
	Element  string
	Decay    float64
	Statmap  *StatMap
}

type StatMap struct {
	sync.Mutex
	Stats       map[string]Stat
	SortOrder   string
	PurgeMethod string
	MaxLen      int
	Tier        int
	ForceResort bool
	Dirty       map[string]bool
	RateUnit    string
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
func (s BySum) Less(i, j int) bool { return s[i].Sum > s[j].Sum }
func (s BySum) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByAverage) Len() int           { return len(s) }
func (s ByAverage) Less(i, j int) bool { return s[i].Average > s[j].Average }
func (s ByAverage) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s BySeen) Len() int           { return len(s) }
func (s BySeen) Less(i, j int) bool { return s[i].Seen > s[j].Seen }
func (s BySeen) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByMax) Len() int           { return len(s) }
func (s ByMax) Less(i, j int) bool { return s[i].Max > s[j].Max }
func (s ByMax) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByMin) Len() int           { return len(s) }
func (s ByMin) Less(i, j int) bool { return s[i].Min < s[j].Min }
func (s ByMin) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByLastSeen) Len() int           { return len(s) }
func (s ByLastSeen) Less(i, j int) bool { return s[i].LastSeen.After(s[j].LastSeen) }
func (s ByLastSeen) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s ByDecay) Len() int           { return len(s) }
func (s ByDecay) Less(i, j int) bool { return s[i].Decay > s[j].Decay }
func (s ByDecay) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (statmap *StatMap) Decay() {
	c := time.Tick(1 * time.Second)
	n := 0
	for _ = range c {
		if n < 60 {
			n++
		}

		statmap.Lock()
		for m, stat := range statmap.Stats {
			stat.Decay = (1.0/3.0 - stat.Decay) / float64(n)
			statmap.Stats[m] = stat

		}
		statmap.Unlock()
	}
	return
}

func (statmap *StatMap) SetSortOrder(sortOrder string) {
	statmap.SortOrder = sortOrder
	statmap.ForceResort = true
	return
}

func (statmap *StatMap) SetTier(tier int) {
	statmap.Tier = tier
	statmap.ForceResort = true
	return
}

func (statmap *StatMap) sort() Stats {

	statmap.Lock()
	defer statmap.Unlock()

	var s Stats
	for _, stat := range statmap.Stats {
		s = append(s, stat)
	}
	s.sort(statmap.SortOrder)
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
	case "percentage":
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

func (statmap *StatMap) Purge() (purged bool) {
	if statmap.MaxLen == -1 {
		return false
	}

	var s Stats
	for _, stat := range statmap.Stats {
		s = append(s, stat)
	}
	s.sort(statmap.PurgeMethod)

	statmap.Lock()
	defer statmap.Unlock()

	if len(s) > statmap.MaxLen {
		s = s[statmap.MaxLen:len(s)]
		for _, stat := range s {
			delete(statmap.Stats, stat.Element)
		}
		return true
	}

	return false
}

func (statmap *StatMap) FastSort() Stats {

	n := statmap.Tier

	var s Stats

	if statmap.ForceResort {
		statmap.Dirty = make(map[string]bool)
		s = statmap.sort()
	} else {
		statmap.Lock()
		for element := range statmap.Dirty {
			s = append(s, statmap.Stats[element])
		}
		statmap.Unlock()
		s.sort(statmap.SortOrder)
	}

	if len(s) > n {
		s = s[:n]
	}

	statmap.Dirty = make(map[string]bool)
	for _, stat := range s {
		statmap.Dirty[stat.Element] = true
	}

	return s
}

func (statmap *StatMap) UpdateElement(num float64, element string) (err error) {

	statmap.Lock()
	defer statmap.Unlock()

	stat, ok := statmap.Stats[element]

	if !ok {
		stat = Stat{}
	}

	max := stat.Max
	min := stat.Min
	if num > max {
		max = num
	}
	if num < min {
		min = num
	}

	statmap.Stats[element] = Stat{
		Sum:      stat.Sum + num,
		Average:  ((stat.Average*float64(stat.Seen) + num) / (float64(stat.Seen) + 1)),
		Seen:     stat.Seen + 1,
		Element:  element,
		Min:      min,
		Max:      max,
		LastSeen: time.Now(),
		Decay:    stat.Decay + 1,
		Statmap:  statmap,
	}
	statmap.Dirty[element] = true
	return
}

func (s *Stat) GetRate(startTime time.Time) float64 {
	unit := s.Statmap.RateUnit
	now := time.Since(startTime)
	var d float64
	switch unit {
	case "hour":
		d = now.Hours()
	case "minute":
		d = now.Minutes()
	case "second":
		d = now.Seconds()
	}
	return float64(s.Seen) / math.Ceil(d)
}

func (s *Stat) GetPercentage() float64 {
	return float64(s.Seen) / float64(len(s.Statmap.Stats)) * 100
}
