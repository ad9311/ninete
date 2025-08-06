package server

import (
	"net/http"
)

// GetHealthz handles the /healthz route and returns a 204 No Content response if the server is up.
func (s *Server) GetHealthz(w http.ResponseWriter, _ *http.Request) {
	writeNoContent(w)
}

// GetReadyz handles the /readyz route and returns server, system, and database stats if ready.
func (s *Server) GetReadyz(w http.ResponseWriter, r *http.Request) {
	if err := s.serviceStore.PingDB(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, internalErrorCode, err)

		return
	}

	dbStats := s.serviceStore.GetPoolStats()
	appStats := s.serviceStore.GetProgramStats()
	sysStas, err := s.serviceStore.GetSystemStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, internalErrorCode, err)
	}

	write(w, http.StatusOK,
		Data{
			"environment": s.config.Env,
			"program": Data{
				"allocatedMemory":        appStats.AllocatedMemory,
				"totalAllocatedMemory":   appStats.TotalAllocatedMemory,
				"systemProgramMemory":    appStats.SystemProgramMemory,
				"garbageCollectorCycles": appStats.GCCycles,
			},
			"system": Data{
				"totalMemory": sysStas.TotalVirtualMemory,
				"freeMemory":  sysStas.FreeVirtualMemory,
				"memoryUsage": sysStas.UsedMemory,
			},
			"database": Data{
				"maxConnections":   dbStats.MaxConns(),
				"totalConnections": dbStats.TotalConns(),
				"idleConnections":  dbStats.IdleConns(),
			},
		},
	)
}
