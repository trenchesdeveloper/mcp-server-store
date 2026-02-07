package mcp

// MCP Protocol version
const ProtocolVersion = "2025-11-25"

// ---- Capability types ----

// ServerCapabilities describes what the MCP server supports.
type ServerCapabilities struct {
	Tools     *ToolCapability     `json:"tools,omitempty"`
	Resources *ResourceCapability `json:"resources,omitempty"`
	Prompts   *PromptCapability   `json:"prompts,omitempty"`
	Logging   *LoggingCapability  `json:"logging,omitempty"`
}

type ToolCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourceCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type PromptCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type LoggingCapability struct{}

// ClientCapabilities describes what the MCP client supports.
type ClientCapabilities struct {
	Experimental map[string]any      `json:"experimental,omitempty"`
	Sampling     *SamplingCapability `json:"sampling,omitempty"`
}

type SamplingCapability struct{}

// ---- Implementation ----

// Implementation identifies a client or server.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ---- Initialize ----

// InitializeParams are sent by the client in the "initialize" request.
type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// InitializeResult is returned by the server in response to "initialize".
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ClientInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

// ---- Ping ----

// PingResult is returned by the server in response to "ping".
type PingResult struct{}

// ---- Pagination ----

// Cursor is an opaque token used to represent a pagination position.
type Cursor string

// PaginatedRequest contains optional cursor-based pagination fields.
type PaginatedRequest struct {
	Cursor *Cursor `json:"cursor,omitempty"`
}

// PaginatedResult contains an optional cursor pointing to the next page.
type PaginatedResult struct {
	NextCursor *Cursor `json:"nextCursor,omitempty"`
}

// ---- Content types ----

// Content represents a content block in an MCP response.
type Content struct {
	Type     string `json:"type"`               // "text", "image", "resource"
	Text     string `json:"text,omitempty"`     // for type "text"
	MimeType string `json:"mimeType,omitempty"` // for type "image"
	Data     string `json:"data,omitempty"`     // for type "image" (base64)
	URI      string `json:"uri,omitempty"`      // for type "resource"
}

// NewTextContent creates a text content block.
func NewTextContent(text string) Content {
	return Content{
		Type: "text",
		Text: text,
	}
}

// NewImageContent creates an image content block with base64-encoded data.
func NewImageContent(mimeType, base64Data string) Content {
	return Content{
		Type:     "image",
		MimeType: mimeType,
		Data:     base64Data,
	}
}

// NewErrorContent creates a text content block marked as an error.
func NewErrorContent(text string) (Content, bool) {
	return Content{
		Type: "text",
		Text: text,
	}, true
}
