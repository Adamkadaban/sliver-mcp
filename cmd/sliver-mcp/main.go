package main

import (
	"flag"
	"log"

	"github.com/adamkadaban/sliver-mcp/internal/server"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to the configuration file")
	var transport string
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or sse)")
	flag.Parse()

	mcpServer := server.NewSliverMCPServer(configPath)

	if transport == "sse" {
		sseServer := mcpgoserver.NewSSEServer(mcpServer, mcpgoserver.WithBaseURL("http://localhost:8080"))
		log.Printf("SSE server listening on :8080")
		if err := sseServer.Start(":8080"); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := mcpgoserver.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}