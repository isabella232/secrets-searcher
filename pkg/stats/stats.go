package stats

import (
	"sort"
	"sync"
	"time"
)

type (
	Stats struct {
		AppStartTime         time.Time
		AppEndTime           time.Time
		SearchStartTime      time.Time
		SearchEndTime        time.Time
		CommitsSearchedCount int64
		SecretsFoundCount    int64
		RepoDurations        DurationStats
		CommitDurations      DurationStats
		FileChangeDurations  DurationStats
		FileTypeDurations    DurationStats
	}
	DurationStats struct {
		stats []*DurationStat
		mu    sync.Mutex
	}
	DurationStat struct {
		Item string
		Dur  time.Duration
	}
)

func New() *Stats {
	return &Stats{
		RepoDurations:       NewAggregatedDurationStats(),
		CommitDurations:     NewUniqueDurationStats(),
		FileChangeDurations: NewUniqueDurationStats(),
		FileTypeDurations:   NewAggregatedDurationStats(),
	}
}

//
// DurationStats

const limit = 10

func NewUniqueDurationStats() DurationStats {
	return DurationStats{stats: make([]*DurationStat, limit)}
}

func NewAggregatedDurationStats() DurationStats {
	return DurationStats{}
}

func (ss *DurationStats) Stats() (result []*DurationStat) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.stats == nil {
		return nil
	}

	stats := *&ss.stats

	// Unique durations will already be sorted
	sort.Slice(stats, func(i, j int) bool {
		return stats[i] != nil && stats[j] != nil && stats[i].Dur > stats[j].Dur
	})

	for i, stat := range stats {
		if stat == nil || i == limit {
			break
		}
		result = append(result, stat)
	}

	return
}

func (ss *DurationStats) SubmitUniqueDuration(dur time.Duration, item string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	for i, stat := range ss.stats {
		if stat == nil {
			ss.stats[i] = &DurationStat{Item: item, Dur: dur}
			return
		}
		if stat.Dur < dur {
			copy(ss.stats[i+1:], ss.stats[i:])
			ss.stats[i] = &DurationStat{Item: item, Dur: dur}
			return
		}
	}
}

func (ss *DurationStats) SubmitAggregatedDuration(dur time.Duration, item string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	for i, stat := range ss.stats {
		if stat.Item == item {
			ss.stats[i].Dur = stat.Dur + dur
			return
		}
	}

	ss.stats = append(ss.stats, &DurationStat{Item: item, Dur: dur})
}
