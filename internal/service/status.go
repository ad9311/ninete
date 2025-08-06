package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shirou/gopsutil/v4/mem"
)

// ProgramStat contains statistics about the current Go program's memory usage and GC cycles.
type ProgramStat struct {
	AllocatedMemory      string // Currently allocated memory (MiB)
	TotalAllocatedMemory string // Total memory allocated since start (MiB)
	SystemProgramMemory  string // System memory obtained for the program (MiB)
	GCCycles             uint32 // Number of completed GC cycles
}

// SystemStat contains statistics about the system's virtual memory usage.
type SystemStat struct {
	TotalVirtualMemory string // Total virtual memory (MiB)
	FreeVirtualMemory  string // Free virtual memory (MiB)
	UsedMemory         string // Percentage of used memory
}

// PingDB checks database connectivity by pinging the current database. Returns an error if the ping fails.
func (s *Store) PingDB(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		return errs.WrapErrorWithMessage("failed to ping database", err)
	}

	return nil
}

// GetPoolStats returns statistics about the current database connection pool.
func (s *Store) GetPoolStats() *pgxpool.Stat {
	return s.db.Stat()
}

// GetProgramStats returns current memory and GC statistics for the running Go program in a readable format.
func (s *Store) GetProgramStats() ProgramStat {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ProgramStat{
		AllocatedMemory:      fmt.Sprintf("%d MiB", m.Alloc/1024/1024),
		TotalAllocatedMemory: fmt.Sprintf("%d MiB", m.TotalAlloc/1024/1024),
		SystemProgramMemory:  fmt.Sprintf("%d MiB", m.Sys/1024/1024),
		GCCycles:             m.NumGC,
	}
}

// GetSystemStats returns system virtual memory statistics in a readable format. Returns an error if stats cannot be retrieved.
func (s *Store) GetSystemStats() (SystemStat, error) {
	var sysStats SystemStat

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return sysStats, err
	}

	return SystemStat{
		TotalVirtualMemory: fmt.Sprintf("%d MiB", vmStat.Total/1024/1024),
		FreeVirtualMemory:  fmt.Sprintf("%d MiB", vmStat.Free/1024/1024),
		UsedMemory:         fmt.Sprintf("%.2f %%", vmStat.UsedPercent),
	}, nil
}
