package logic

import (
	"fmt"
	"runtime"

	"github.com/ad9311/ninete/internal/repo"
)

type AppStats struct {
	Version     string           `json:"version"`
	ENV         string           `json:"environment"`
	DBConnStats repo.DBConnStats `json:"database"`
	MemStats    MemStats         `json:"memory"`
}

type MemStats struct {
	Allocated string `json:"allocated"`
	System    string `json:"system"`
}

func (s *Store) AppStatus() (AppStats, error) {
	var stats AppStats

	dbStats, err := s.queries.CheckDBStatus()
	if err != nil {
		return stats, err
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memStats := MemStats{
		Allocated: fmt.Sprintf("%d MiB", bToMb(m.Alloc)),
		System:    fmt.Sprintf("%d MiB", bToMb(m.Sys)),
	}

	stats = AppStats{
		Version:     "TODO",
		ENV:         s.app.ENV,
		DBConnStats: dbStats,
		MemStats:    memStats,
	}

	return stats, nil
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
