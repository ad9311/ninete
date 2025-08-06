package server_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type readyzBody struct {
	Environment string `json:"environment"`
	Program     struct {
		AllocatedMemory        string  `json:"allocatedMemory"`
		TotalAllocatedMemory   string  `json:"totalAllocatedMemory"`
		SystemProgramMemory    string  `json:"systemProgramMemory"`
		GarbageCollectorCycles float64 `json:"garbageCollectorCycles"`
	} `json:"program"`
	System struct {
		TotalMemory string `json:"totalMemory"`
		FreeMemory  string `json:"freeMemory"`
		MemoryUsage string `json:"memoryUsage"`
	} `json:"system"`
	Database struct {
		MaxConnections   float64 `json:"maxConnections"`
		TotalConnections float64 `json:"totalConnections"`
		IdleConnections  float64 `json:"idleConnections"`
	} `json:"database"`
}

func TestGetHealthz(t *testing.T) {
	fs := newFactoryServer(t)
	res, req := newHTTPTest(factoryHTTP{
		method: http.MethodGet,
		target: "/healthz",
	})

	fs.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestGetReadyz(t *testing.T) {
	fs := newFactoryServer(t)
	res, req := newHTTPTest(factoryHTTP{
		method: http.MethodGet,
		target: "/readyz",
	})

	fs.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var resBody factoryResponse
	decodeJSONBody(t, res, &resBody)

	var data readyzBody
	dataToStruct(t, resBody.Data, &data)

	require.NotEmpty(t, data.Environment)
	require.Equal(t, data.Environment, "test")
	require.NotEmpty(t, data.Program.AllocatedMemory)
	require.NotEmpty(t, data.Program.TotalAllocatedMemory)
	require.NotEmpty(t, data.Program.SystemProgramMemory)
	require.IsType(t, data.Program.GarbageCollectorCycles, float64(0))
	require.NotEmpty(t, data.System.TotalMemory)
	require.NotEmpty(t, data.System.FreeMemory)
	require.NotEmpty(t, data.System.MemoryUsage)
	require.NotEmpty(t, data.Database.MaxConnections)
	require.NotEmpty(t, data.Database.TotalConnections)
	require.NotEmpty(t, data.Database.IdleConnections)
}
