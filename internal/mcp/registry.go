package mcp

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/trenchesdeveloper/mcp-server-store/internal/jsonrpc"
)

// e.g list Products, get Product by ID, create Product, update Product, delete Product
// ToolHandler is a function that executes a tool and returns the result.
type ToolHandler func(arguments map[string]interface{}) (*ToolCallResult, error)

// ResourceHandler is a function that reads a resource and returns its contents.
type ResourceHandler func(uri string) (*ReadResourceResult, error)

// PromptHandler is a function that resolves a prompt with the given arguments.
type PromptHandler func(arguments map[string]string) (*GetPromptResult, error)

// Registry is the central MCP server that registers tools, resources, and prompts,
// and wires them up as JSON-RPC method handlers.
type Registry struct {
	serverInfo   ClientInfo
	capabilities ServerCapabilities
	instructions string

	tools        map[string]Tool
	toolHandlers map[string]ToolHandler

	resources        map[string]Resource
	resourceHandlers map[string]ResourceHandler

	prompts        map[string]Prompt
	promptHandlers map[string]PromptHandler

	logger *logrus.Logger
	mu     sync.RWMutex
}

// NewRegistry creates a new MCP registry with the given server info and instructions.
func NewRegistry(serverInfo ClientInfo, instructions string, logger *logrus.Logger) *Registry {
	return &Registry{
		serverInfo:       serverInfo,
		instructions:     instructions,
		tools:            make(map[string]Tool),
		toolHandlers:     make(map[string]ToolHandler),
		resources:        make(map[string]Resource),
		resourceHandlers: make(map[string]ResourceHandler),
		prompts:          make(map[string]Prompt),
		promptHandlers:   make(map[string]PromptHandler),
		logger:           logger,
	}
}

// ---- Registration methods ----

// RegisterTool adds a tool and its handler to the registry.
func (r *Registry) RegisterTool(tool Tool, handler ToolHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
	r.toolHandlers[tool.Name] = handler
	r.logger.WithField("tool", tool.Name).Info("Registered tool")
}

// RegisterResource adds a resource and its handler to the registry.
func (r *Registry) RegisterResource(resource Resource, handler ResourceHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources[resource.URI] = resource
	r.resourceHandlers[resource.URI] = handler
	r.logger.WithField("resource", resource.URI).Info("Registered resource")
}

// RegisterPrompt adds a prompt and its handler to the registry.
func (r *Registry) RegisterPrompt(prompt Prompt, handler PromptHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prompts[prompt.Name] = prompt
	r.promptHandlers[prompt.Name] = handler
	r.logger.WithField("prompt", prompt.Name).Info("Registered prompt")
}

// ---- Wire up to JSON-RPC server ----

// RegisterHandlers registers all MCP protocol methods on the given JSON-RPC server.
func (r *Registry) RegisterHandlers(server *jsonrpc.Server) {
	// Build capabilities based on what is registered
	r.capabilities = r.buildCapabilities()

	server.RegisterMethod(MethodInitialize, r.handleInitialize)
	server.RegisterMethod(MethodPing, r.handlePing)

	if r.capabilities.Tools != nil {
		server.RegisterMethod(MethodToolsList, r.handleToolsList)
		server.RegisterMethod(MethodToolsCall, r.handleToolsCall)
	}

	if r.capabilities.Resources != nil {
		server.RegisterMethod(MethodResourcesList, r.handleResourcesList)
		server.RegisterMethod(MethodResourcesRead, r.handleResourcesRead)
	}

	if r.capabilities.Prompts != nil {
		server.RegisterMethod(MethodPromptsList, r.handlePromptsList)
		server.RegisterMethod(MethodPromptsGet, r.handlePromptsGet)
	}

	// Notifications (no response expected)
	server.RegisterMethod(NotificationInitialized, r.handleInitializedNotification)
}

// ---- Capability builder ----

func (r *Registry) buildCapabilities() ServerCapabilities {
	caps := ServerCapabilities{
		Logging: &LoggingCapability{},
	}

	if len(r.tools) > 0 {
		caps.Tools = &ToolCapability{ListChanged: false}
	}
	if len(r.resources) > 0 {
		caps.Resources = &ResourceCapability{Subscribe: false, ListChanged: false}
	}
	if len(r.prompts) > 0 {
		caps.Prompts = &PromptCapability{ListChanged: false}
	}

	return caps
}

// ---- Handler implementations ----

func (r *Registry) handleInitialize(params json.RawMessage) (interface{}, *jsonrpc.Error) {
	var req InitializeRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid initialize params", err.Error())
	}

	r.logger.WithFields(logrus.Fields{
		"client":          req.ClientInfo.Name,
		"clientVersion":   req.ClientInfo.Version,
		"protocolVersion": req.ProtocolVersion,
	}).Info("Client initializing")

	return &InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities:    r.capabilities,
		ServerInfo:      r.serverInfo,
		Instructions:    r.instructions,
	}, nil
}

func (r *Registry) handlePing(_ json.RawMessage) (interface{}, *jsonrpc.Error) {
	return &PingResult{}, nil
}

func (r *Registry) handleInitializedNotification(_ json.RawMessage) (interface{}, *jsonrpc.Error) {
	r.logger.Info("Client initialized successfully")
	return nil, nil
}

// ---- Tool handlers ----

func (r *Registry) handleToolsList(_ json.RawMessage) (interface{}, *jsonrpc.Error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.logger.WithField("count", len(r.tools)).Info("Listing tools")

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return &ToolListResult{Tools: tools}, nil
}

func (r *Registry) handleToolsCall(params json.RawMessage) (interface{}, *jsonrpc.Error) {
	var req ToolCallParams
	if err := json.Unmarshal(params, &req); err != nil {
		r.logger.WithError(err).Error("Failed to parse tool call params")
		return nil, jsonrpc.NewInvalidParamsError("Invalid tool call params", err.Error())
	}

	r.logger.WithFields(logrus.Fields{
		"tool":      req.Name,
		"arguments": req.Arguments,
	}).Info("Calling tool")

	r.mu.RLock()
	handler, ok := r.toolHandlers[req.Name]
	r.mu.RUnlock()

	if !ok {
		r.logger.WithField("tool", req.Name).Warn("Tool not found")
		return nil, jsonrpc.NewInvalidParamsError(
			fmt.Sprintf("Tool '%s' not found", req.Name), nil,
		)
	}

	result, err := handler(req.Arguments)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"tool":  req.Name,
			"error": err.Error(),
		}).Error("Tool execution failed")
		// Return the error as a tool result with isError=true, not a JSON-RPC error.
		return &ToolCallResult{
			Content: []Content{NewTextContent(err.Error())},
			IsError: true,
		}, nil
	}

	r.logger.WithField("tool", req.Name).Info("Tool executed successfully")

	return result, nil
}

// ---- Resource handlers ----

func (r *Registry) handleResourcesList(_ json.RawMessage) (interface{}, *jsonrpc.Error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resources := make([]Resource, 0, len(r.resources))
	for _, res := range r.resources {
		resources = append(resources, res)
	}

	return &ListResourcesResult{Resources: resources}, nil
}

func (r *Registry) handleResourcesRead(params json.RawMessage) (interface{}, *jsonrpc.Error) {
	var req ReadResourceParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid resource read params", err.Error())
	}

	r.mu.RLock()
	handler, ok := r.resourceHandlers[req.URI]
	r.mu.RUnlock()

	if !ok {
		return nil, jsonrpc.NewInvalidParamsError(
			fmt.Sprintf("Resource '%s' not found", req.URI), nil,
		)
	}

	result, err := handler(req.URI)
	if err != nil {
		return nil, jsonrpc.NewInternalError("Failed to read resource", err.Error())
	}

	return result, nil
}

// ---- Prompt handlers ----

func (r *Registry) handlePromptsList(_ json.RawMessage) (interface{}, *jsonrpc.Error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prompts := make([]Prompt, 0, len(r.prompts))
	for _, p := range r.prompts {
		prompts = append(prompts, p)
	}

	return &ListPromptsResult{Prompts: prompts}, nil
}

func (r *Registry) handlePromptsGet(params json.RawMessage) (interface{}, *jsonrpc.Error) {
	var req GetPromptParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewInvalidParamsError("Invalid prompt get params", err.Error())
	}

	r.mu.RLock()
	handler, ok := r.promptHandlers[req.Name]
	r.mu.RUnlock()

	if !ok {
		return nil, jsonrpc.NewInvalidParamsError(
			fmt.Sprintf("Prompt '%s' not found", req.Name), nil,
		)
	}

	result, err := handler(req.Arguments)
	if err != nil {
		return nil, jsonrpc.NewInternalError("Failed to get prompt", err.Error())
	}

	return result, nil
}
