package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/trenchesdeveloper/mcp-server-store/configs"
	"github.com/trenchesdeveloper/mcp-server-store/internal/client"
	"github.com/trenchesdeveloper/mcp-server-store/internal/mcp"
)

func main() {
	// Load configuration
	cfg := configs.LoadConfig()

	// Configure logging
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.DebugLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	// Log to stderr so stdout stays clean for JSON-RPC
	logger.SetOutput(os.Stderr)

	logger.Info("Starting MCP Server...")

	// Create HTTP client for the ecommerce API
	httpClient := client.NewRestClient(cfg.APIURL, cfg.AuthToken, logger)

	// Create the MCP server
	server := mcp.NewServer(
		"mcp-server-store",
		"0.1.0",
		logger,
		mcp.WithInstructions("A store management MCP server."),
		mcp.WithHTTPClient(httpClient),
	)

	logger.WithField("tools", len(server.ListTools())).Info("Registered tools")

	// Start serving over stdio
	if err := server.ServeStdio(); err != nil {
		logger.WithError(err).Fatal("Server exited with error")
	}
}
