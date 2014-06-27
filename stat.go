package main

import "sort"
import "time"

type Stat struct {
	sum       float64
	average   float64
	min       float64
	max       float64
	last_seen time.Time
	seen      int
	element   string
}

type Stats []Stat

type BySum []Stat
type ByAverage []Stat
type BySeen []Stat
type ByMax []Stat
type ByMin []Stat
type ByLastSeen []Stat

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
func (s ByLastSeen) Less(i, j int) bool { return s[i].last_seen.After(s[j].last_seen) }
func (s ByLastSeen) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func statsort(sort_order string, m map[string]Stat) Stats {
	var s Stats
	for _, stat := range m {
		s = append(s, stat)
	}
	switch sort_order {
	case "sum":
		sort.Sort(BySum(s))
	case "average":
		sort.Sort(ByAverage(s))
	case "seen":
		sort.Sort(BySeen(s))
	case "max":
		sort.Sort(ByMax(s))
	case "min":
		sort.Sort(ByMin(s))
	case "last_seen":
		sort.Sort(ByLastSeen(s))
	}
	return s
}

func elementsort(sort_order string, m map[string]Stat) []string {
	stats := statsort(sort_order, m)
	var keys []string
	for _, stat := range stats {
		keys = append(keys, stat.element)
	}
	return keys
}

func purge_stats(purge_method string, keep int, m map[string]Stat) (purged bool) {
	keys := elementsort(purge_method, m)
	if len(keys) > keep {
		keys = keys[keep:len(keys)]
		for _, key := range keys {
			delete(m, key)
		}
		return true
	}

	return false
}
