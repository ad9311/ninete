// Command ninete-mcp is a local MCP server that lets an MCP client (Claude
// Code, Claude Desktop) work with a Ninete account through its /api/*, using a
// personal access token. It speaks MCP over stdio: the client launches it and
// passes NINETE_URL, NINETE_TOKEN and optionally NINETE_TZ in its environment.
//
// stdout is the protocol channel. Nothing else may write to it; diagnostics go
// to stderr.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ad9311/ninete-mcp/internal/api"
	"github.com/ad9311/ninete-mcp/internal/config"
	"github.com/ad9311/ninete-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is stamped with -X main.version at build time (make build-mcp).
var version = "dev" //nolint:gochecknoglobals // set by the linker

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ninete-mcp:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := mcp.NewServer(&mcp.Implementation{Name: "ninete", Version: version}, &mcp.ServerOptions{
		Instructions: "Tools for the owner's Ninete expense tracker. Amounts are decimal strings in the " +
			"account's currency. An expense's billed_month (YYYY-MM) is the month it counts toward. " +
			"Nothing can be deleted through these tools; the owner does that in the app.",
	})

	tools.Register(server, tools.Deps{
		API:      api.New(cfg.BaseURL, cfg.Token, version),
		Location: cfg.Location,
		Now:      time.Now,
	})

	return server.Run(ctx, &mcp.StdioTransport{})
}
