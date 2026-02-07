package products

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/trenchesdeveloper/mcp-server-store/internal/client"
	"github.com/trenchesdeveloper/mcp-server-store/internal/mcp"
)

// ProductToolSet groups all product-related tools and shares the HTTP client.
type ProductToolSet struct {
	httpClient *client.RestClient
	logger     *logrus.Logger
}

// NewProductToolSet creates a new ProductToolSet with the given HTTP client and logger.
func NewProductToolSet(httpClient *client.RestClient, logger *logrus.Logger) *ProductToolSet {
	return &ProductToolSet{httpClient: httpClient, logger: logger}
}

// ---- List Products ----

// ListTool returns the tool definition for listing products.
func (p *ProductToolSet) ListTool() mcp.Tool {
	return mcp.Tool{
		Name:        "list_products",
		Description: "Lists products from the ecommerce store. Supports optional pagination with page and limit parameters.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"page": {
					Type:        "string",
					Description: "Page number for pagination (default: 1)",
				},
				"limit": {
					Type:        "string",
					Description: "Number of products per page (default: 10)",
				},
			},
		},
	}
}

// ListHandler returns a handler that fetches products from the ecommerce API.
func (p *ProductToolSet) ListHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		p.logger.WithField("arguments", arguments).Info("Listing products")

		params := map[string]string{}

		if page, ok := arguments["page"].(string); ok && page != "" {
			params["page"] = page
		}
		if limit, ok := arguments["limit"].(string); ok && limit != "" {
			params["limit"] = limit
		}

		body, err := p.httpClient.Get("/products", params)
		if err != nil {
			p.logger.WithError(err).Error("Failed to list products")
			return nil, fmt.Errorf("failed to list products: %w", err)
		}

		var resp ProductResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			p.logger.WithError(err).Error("Failed to parse products response")
			return nil, fmt.Errorf("failed to parse products response: %w", err)
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Found %d products\n\n", len(resp.Data))

		for i, product := range resp.Data {
			fmt.Fprintf(&sb, "%d. %s\n", i+1, formatProduct(product))
		}

		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: sb.String(),
				},
			},
		}, nil
	}
}

func formatProduct(p Product) string {
	name := p.Name
	price := p.Price
	id := p.ID

	return fmt.Sprintf("**%s** (ID: %d) - $%.2f", name, id, price)
}

// ---- Search Products ----

// SearchTool returns the tool definition for searching products.
func (p *ProductToolSet) SearchTool() mcp.Tool {
	return mcp.Tool{
		Name:        "search_products",
		Description: "Full-text search products by name, SKU, and description with optional filters for category, price range, and pagination.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"q": {
					Type:        "string",
					Description: "Search query (searches name, SKU, and description)",
				},
				"page": {
					Type:        "string",
					Description: "Page number for pagination (default: 1)",
				},
				"limit": {
					Type:        "string",
					Description: "Number of results per page (default: 10)",
				},
				"category_id": {
					Type:        "string",
					Description: "Filter by category ID",
				},
				"min_price": {
					Type:        "string",
					Description: "Minimum price filter",
				},
				"max_price": {
					Type:        "string",
					Description: "Maximum price filter",
				},
			},
			Required: []string{"q"},
		},
	}
}

// SearchHandler returns a handler that searches products via the ecommerce API.
func (p *ProductToolSet) SearchHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		p.logger.WithField("arguments", arguments).Info("Searching products")

		params := map[string]string{}

		for _, key := range []string{"q", "page", "limit", "category_id", "min_price", "max_price"} {
			if val, ok := arguments[key].(string); ok && val != "" {
				params[key] = val
			}
		}

		body, err := p.httpClient.Get("/products/search", params)
		if err != nil {
			p.logger.WithError(err).Error("Failed to search products")
			return nil, fmt.Errorf("failed to search products: %w", err)
		}

		var resp ProductResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			p.logger.WithError(err).Error("Failed to parse search response")
			return nil, fmt.Errorf("failed to parse search response: %w", err)
		}

		p.logger.WithField("count", len(resp.Data)).Info("Product search completed")

		var sb strings.Builder
		fmt.Fprintf(&sb, "Found %d products matching '%s'\n\n", len(resp.Data), params["q"])

		for i, product := range resp.Data {
			fmt.Fprintf(&sb, "%d. %s\n", i+1, formatProduct(product))
		}

		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: sb.String(),
				},
			},
		}, nil
	}
}

// ---- Product Details ----

// GetDetailTool returns the tool definition for getting a single product by ID.
func (p *ProductToolSet) GetDetailTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_product",
		Description: "Gets detailed information about a specific product by its ID, including name, description, price, stock, category, and images.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"id": {
					Type:        "string",
					Description: "The product ID",
				},
			},
			Required: []string{"id"},
		},
	}
}

// GetDetailHandler returns a handler that fetches a product by ID.
func (p *ProductToolSet) GetDetailHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		id, ok := arguments["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("product id is required")
		}

		p.logger.WithField("id", id).Info("Getting product details")

		body, err := p.httpClient.Get("/products/"+id, nil)
		if err != nil {
			p.logger.WithError(err).Error("Failed to get product details")
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		var resp ProductDetailResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			p.logger.WithError(err).Error("Failed to parse product response")
			return nil, fmt.Errorf("failed to parse product response: %w", err)
		}

		p.logger.WithField("product", resp.Data.Name).Info("Product details retrieved")

		var sb strings.Builder
		fmt.Fprintf(&sb, "**%s** (ID: %d)\n", resp.Data.Name, resp.Data.ID)
		fmt.Fprintf(&sb, "- SKU: %s\n", resp.Data.SKU)
		fmt.Fprintf(&sb, "- Price: $%.2f\n", resp.Data.Price)
		fmt.Fprintf(&sb, "- Stock: %d\n", resp.Data.Stock)
		fmt.Fprintf(&sb, "- Category: %s\n", resp.Data.Category.Name)
		fmt.Fprintf(&sb, "- Active: %v\n", resp.Data.IsActive)
		fmt.Fprintf(&sb, "- Description: %s\n", resp.Data.Description)

		if len(resp.Data.Images) > 0 {
			fmt.Fprintf(&sb, "\nImages:\n")
			for _, img := range resp.Data.Images {
				fmt.Fprintf(&sb, "  - %s (%s)\n", img.URL, img.AltText)
			}
		}

		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: sb.String(),
				},
			},
		}, nil
	}
}

