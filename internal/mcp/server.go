package mcp

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/trenchesdeveloper/mcp-server-store/internal/client"
	"github.com/trenchesdeveloper/mcp-server-store/internal/jsonrpc"
)

// Server is the top-level MCP server. It owns the JSON-RPC server and the
// registry, providing a simple API to register tools/resources/prompts and
// start serving over stdio.
type Server struct {
	rpcServer    *jsonrpc.Server
	registry     *Registry
	logger       *logrus.Logger
	serverInfo   ClientInfo
	instructions string
	capabilities ServerCapabilities
	httpClient   *client.RestClient
}

// ServerOption is a functional option for configuring the MCP Server.
type ServerOption func(*Server)

// WithInstructions sets the natural-language instructions returned during initialize.
func WithInstructions(instructions string) ServerOption {
	return func(s *Server) {
		s.instructions = instructions
	}
}

// WithHTTPClient sets the HTTP client used by tools to call the ecommerce API.
func WithHTTPClient(httpClient *client.RestClient) ServerOption {
	return func(s *Server) {
		s.httpClient = httpClient
	}
}

// NewServer creates a new MCP server with the given name, version, and options.
func NewServer(name, version string, logger *logrus.Logger, opts ...ServerOption) *Server {
	serverInfo := ClientInfo{
		Name:    name,
		Version: version,
	}

	s := &Server{
		rpcServer:  jsonrpc.NewServer(logger),
		registry:   NewRegistry(serverInfo, "", logger),
		logger:     logger,
		serverInfo: serverInfo,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ---- Registration convenience methods ----

// RegisterTool registers a tool with the MCP server.
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.registry.RegisterTool(tool, handler)
}

// RegisterResource registers a resource with the MCP server.
func (s *Server) RegisterResource(resource Resource, handler ResourceHandler) {
	s.registry.RegisterResource(resource, handler)
}

// RegisterPrompt registers a prompt with the MCP server.
func (s *Server) RegisterPrompt(prompt Prompt, handler PromptHandler) {
	s.registry.RegisterPrompt(prompt, handler)
}

// ListTools returns all registered tools.
func (s *Server) ListTools() []Tool {
	s.registry.mu.RLock()
	defer s.registry.mu.RUnlock()
	tools := make([]Tool, 0, len(s.registry.tools))
	for _, tool := range s.registry.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ---- Handler registration ----

// registerHandlers wires up all MCP protocol methods on the JSON-RPC server.
func (s *Server) registerHandlers() {
	// Build capabilities based on registered tools/resources/prompts
	s.capabilities = s.registry.buildCapabilities()

	// Core protocol methods
	s.rpcServer.RegisterMethod(MethodInitialize, s.handleInitialize)
	s.rpcServer.RegisterMethod(MethodPing, s.handlePing)

	// Tool methods
	if s.capabilities.Tools != nil {
		s.rpcServer.RegisterMethod(MethodToolsList, s.registry.handleToolsList)
		s.rpcServer.RegisterMethod(MethodToolsCall, s.registry.handleToolsCall)
	}

	// Resource methods
	if s.capabilities.Resources != nil {
		s.rpcServer.RegisterMethod(MethodResourcesList, s.registry.handleResourcesList)
		s.rpcServer.RegisterMethod(MethodResourcesRead, s.registry.handleResourcesRead)
	}

	// Prompt methods
	if s.capabilities.Prompts != nil {
		s.rpcServer.RegisterMethod(MethodPromptsList, s.registry.handlePromptsList)
		s.rpcServer.RegisterMethod(MethodPromptsGet, s.registry.handlePromptsGet)
	}

	// Notifications (no response expected)
	s.rpcServer.RegisterMethod(NotificationInitialized, s.handleInitializedNotification)

	// Logging
	s.rpcServer.RegisterMethod(MethodLoggingSetLevel, s.handleSetLogLevel)
}

// ---- Handler implementations ----

// handleInitialize handles the "initialize" request from the client.
// It returns the server info, capabilities, protocol version, and instructions.
func (s *Server) handleInitialize(params json.RawMessage) (interface{}, *jsonrpc.Error) {
	var req InitializeRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid initialize params", err.Error())
	}

	s.logger.WithFields(logrus.Fields{
		"client":          req.ClientInfo.Name,
		"clientVersion":   req.ClientInfo.Version,
		"protocolVersion": req.ProtocolVersion,
	}).Info("Client initializing")

	return &InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities:    s.capabilities,
		ServerInfo:      s.serverInfo,
		Instructions:    s.instructions,
	}, nil
}

// handlePing handles the "ping" request.
func (s *Server) handlePing(_ json.RawMessage) (interface{}, *jsonrpc.Error) {
	return &PingResult{}, nil
}

// handleInitializedNotification handles the "notifications/initialized" notification.
func (s *Server) handleInitializedNotification(_ json.RawMessage) (interface{}, *jsonrpc.Error) {
	s.logger.Info("Client initialized successfully")
	return nil, nil
}

// handleSetLogLevel handles the "logging/setLevel" request from the client.
func (s *Server) handleSetLogLevel(params json.RawMessage) (interface{}, *jsonrpc.Error) {
	var req SetLevelParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid logging params", err.Error())
	}

	// Map MCP logging levels to logrus levels
	levelMap := map[LoggingLevel]logrus.Level{
		LogLevelDebug:     logrus.DebugLevel,
		LogLevelInfo:      logrus.InfoLevel,
		LogLevelNotice:    logrus.InfoLevel,
		LogLevelWarning:   logrus.WarnLevel,
		LogLevelError:     logrus.ErrorLevel,
		LogLevelCritical:  logrus.FatalLevel,
		LogLevelAlert:     logrus.FatalLevel,
		LogLevelEmergency: logrus.PanicLevel,
	}

	logrusLevel, ok := levelMap[req.Level]
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("Unknown log level", string(req.Level))
	}

	s.logger.SetLevel(logrusLevel)
	s.logger.WithField("level", req.Level).Info("Log level updated")

	return struct{}{}, nil
}

// ---- Serve ----

// ServeStdio wires up all registered MCP handlers to the JSON-RPC server
// and starts reading from stdin / writing to stdout.
// This method blocks until stdin is closed (EOF) or an error occurs.
func (s *Server) ServeStdio() error {
	s.logger.WithFields(logrus.Fields{
		"server":  s.serverInfo.Name,
		"version": s.serverInfo.Version,
	}).Info("Starting MCP server over stdio")

	// Wire up all MCP protocol methods to the JSON-RPC server.
	s.registerHandlers()

	// Start the stdio read loop.
	return s.rpcServer.ServeStdio()
}

func (s *Server) Start() error {
	return s.ServeStdio()
}
