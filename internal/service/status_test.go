package service_test

import (
	"context"
	"testing"

	"github.com/ad9311/go-api-base/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPingDB(t *testing.T) {
	ctx := context.Background()
	store := service.FactoryStore(t)
	defer store.ClosePool()

	err := store.PingDB(ctx)
	require.Nil(t, err)
}

func TestGetPoolStats(t *testing.T) {
	store := service.FactoryStore(t)
	defer store.ClosePool()

	stats := store.GetPoolStats()
	require.NotNil(t, stats)
}

func TestGetProgramStats(t *testing.T) {
	store := service.FactoryStore(t)
	defer store.ClosePool()

	stats := store.GetProgramStats()
	require.NotEmpty(t, stats.AllocatedMemory)
	require.NotEmpty(t, stats.TotalAllocatedMemory)
	require.NotEmpty(t, stats.SystemProgramMemory)
	require.Positive(t, stats.GCCycles)
}

func TestGetSystemStats(t *testing.T) {
	store := service.FactoryStore(t)
	defer store.ClosePool()

	stats, err := store.GetSystemStats()
	require.NoError(t, err)
	require.NotEmpty(t, stats.TotalVirtualMemory)
	require.NotEmpty(t, stats.FreeVirtualMemory)
	require.NotEmpty(t, stats.UsedMemory)
}
