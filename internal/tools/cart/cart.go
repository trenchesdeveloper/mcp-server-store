package cart

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/trenchesdeveloper/mcp-server-store/internal/client"
	"github.com/trenchesdeveloper/mcp-server-store/internal/mcp"
)

// CartToolSet groups all cart-related tools and shares the HTTP client.
type CartToolSet struct {
	httpClient *client.RestClient
	logger     *logrus.Logger
}

// NewCartToolSet creates a new CartToolSet with the given HTTP client and logger.
func NewCartToolSet(httpClient *client.RestClient, logger *logrus.Logger) *CartToolSet {
	return &CartToolSet{httpClient: httpClient, logger: logger}
}

// ---- Add to Cart ----

// AddToCartTool returns the tool definition for adding a product to the cart.
func (c *CartToolSet) AddToCartTool() mcp.Tool {
	return mcp.Tool{
		Name:        "add_to_cart",
		Description: "Adds a product to the shopping cart. Requires authentication.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"product_id": {
					Type:        "string",
					Description: "The ID of the product to add to the cart",
				},
				"quantity": {
					Type:        "string",
					Description: "The quantity to add (default: 1)",
				},
			},
			Required: []string{"product_id"},
		},
	}
}

// AddToCartHandler returns a handler that adds a product to the cart.
func (c *CartToolSet) AddToCartHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		c.logger.WithField("arguments", arguments).Info("Adding product to cart")

		productIDStr, ok := arguments["product_id"].(string)
		if !ok || productIDStr == "" {
			return nil, fmt.Errorf("product_id is required")
		}

		productID, err := strconv.ParseUint(productIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid product_id: %w", err)
		}

		quantity := 1
		if qtyStr, ok := arguments["quantity"].(string); ok && qtyStr != "" {
			if q, err := strconv.Atoi(qtyStr); err == nil && q > 0 {
				quantity = q
			}
		}

		reqBody := AddToCartRequest{
			ProductID: uint(productID),
			Quantity:  quantity,
		}

		body, err := c.httpClient.WithToken().Post("/cart/items", reqBody)
		if err != nil {
			c.logger.WithError(err).Error("Failed to add product to cart")
			return nil, fmt.Errorf("failed to add to cart: %w", err)
		}

		var resp AddToCartResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			c.logger.WithError(err).Error("Failed to parse cart response")
			return nil, fmt.Errorf("failed to parse cart response: %w", err)
		}

		c.logger.WithFields(logrus.Fields{
			"product_id": productID,
			"quantity":   quantity,
		}).Info("Product added to cart")

		result := fmt.Sprintf("Added %d x product #%d to cart.\nCart ID: %d, Total: $%.2f",
			quantity, productID, resp.Data.ID, resp.Data.Total)

		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				mcp.NewTextContent(result),
			},
		}, nil
	}
}

// ---- View Cart ----

// ViewCartTool returns the tool definition for viewing the cart.
func (c *CartToolSet) ViewCartTool() mcp.Tool {
	return mcp.Tool{
		Name:        "view_cart",
		Description: "Views the current shopping cart contents, including all items and the total. Requires authentication.",
		InputSchema: mcp.InputSchema{
			Type: "object",
		},
	}
}

// ViewCartHandler returns a handler that fetches the current cart.
func (c *CartToolSet) ViewCartHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		c.logger.Info("Viewing cart")

		body, err := c.httpClient.WithToken().Get("/cart", nil)
		if err != nil {
			c.logger.WithError(err).Error("Failed to view cart")
			return nil, fmt.Errorf("failed to view cart: %w", err)
		}

		var resp ViewCartResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			c.logger.WithError(err).Error("Failed to parse cart response")
			return nil, fmt.Errorf("failed to parse cart response: %w", err)
		}

		c.logger.WithField("items", len(resp.Data.CartItems)).Info("Cart retrieved")

		var sb strings.Builder

		if len(resp.Data.CartItems) == 0 {
			sb.WriteString("Your cart is empty.")
		} else {
			fmt.Fprintf(&sb, "Shopping Cart (ID: %d)\n\n", resp.Data.ID)
			for i, item := range resp.Data.CartItems {
				fmt.Fprintf(&sb, "%d. %s (ID: %d) - $%.2f\n",
					i+1, item.Product.Name, item.Product.ID, item.Product.Price)
			}
			fmt.Fprintf(&sb, "\nTotal: $%.2f\n", resp.Data.Total)
		}

		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				mcp.NewTextContent(sb.String()),
			},
		}, nil
	}
}
