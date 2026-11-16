// Package bench measures the store algorithms directly, without HTTP.
package bench

import (
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/vanny-laure-lamorte/CacheLab/internal/store"
)

// runs is the number of measurements; the median is kept.
const runs = 3

const minDuration = 100 * time.Millisecond

// Result is the measure of one mode for one size.
type Result struct {
	Mode          string  `json:"mode"`
	Size          int     `json:"size"`
	CreateNsPerOp float64 `json:"create_ns_per_op"`
	GetNsPerOp    float64 `json:"get_ns_per_op"`
	CreateTotalMs float64 `json:"create_total_ms"`
	GetTotalMs    float64 `json:"get_total_ms"`
	MemoryMB      float64 `json:"memory_mb"`
	DurationMs    float64 `json:"duration_ms"`
}

// Run creates then reads size keys in a new store of the given mode.
// It uses its own store: the cache used by the API is not touched.
func Run(mode string, size int) (Result, error) {
	started := time.Now()
	// Keys are prepared before, so their creation is not measured
	keys := make([]string, size)
	for i := range keys {
		keys[i] = fmt.Sprintf("bench-%d", i)
	}

	createTimes := make([]float64, 0, runs)
	getTimes := make([]float64, 0, runs)

	if _, err := store.New(mode); err != nil {
		return Result{}, err
	}

	store.Logs = false
	defer func() { store.Logs = true }()

	for r := 0; r < runs; r++ {
		// 1. Creation: one round = a new store filled with all the keys
		createTimes = append(createTimes, measure(size, func() {
			s, _ := store.New(mode)
			for _, k := range keys {
				s.CreateKey(k, "v")
			}
		}))

		// 2. Reading: one store, read all the keys again and again
		s, _ := store.New(mode)
		for _, k := range keys {
			s.CreateKey(k, "v")
		}
		getTimes = append(getTimes, measure(size, func() {
			for _, k := range keys {
				s.GetKey(k)
			}
		}))
	}

	memory, err := memoryMB(mode, keys)
	if err != nil {
		return Result{}, err
	}
	createNs := median(createTimes)
	getNs := median(getTimes)

	return Result{
		Mode:          mode,
		Size:          size,
		CreateNsPerOp: createNs,
		GetNsPerOp:    getNs,
		CreateTotalMs: createNs * float64(size) / 1e6,
		GetTotalMs:    getNs * float64(size) / 1e6,
		MemoryMB:      memory,
		DurationMs:    float64(time.Since(started).Microseconds()) / 1000,
	}, nil
}

// memoryMB measures the memory used by a store filled with the keys.
func memoryMB(mode string, keys []string) (float64, error) {
	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	s, err := store.New(mode)
	if err != nil {
		return 0, err
	}
	for _, k := range keys {
		s.CreateKey(k, "v")
	}

	runtime.GC()
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(s) // keep the store alive until the measure is done

	return float64(after.HeapAlloc-before.HeapAlloc) / 1e6, nil
}

// measure repeats round until minDuration has passed,
// and returns the average time of one operation (round does opsPerRound operations).
func measure(opsPerRound int, round func()) float64 {
	rounds := 0
	start := time.Now()
	for time.Since(start) < minDuration {
		round()
		rounds++
	}
	return float64(time.Since(start).Nanoseconds()) / float64(rounds*opsPerRound)
}

// median returns the middle value (less sensitive to a slow run than the average).
func median(values []float64) float64 {
	sort.Float64s(values)
	return values[len(values)/2]
}
