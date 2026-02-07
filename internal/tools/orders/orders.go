package orders

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/trenchesdeveloper/mcp-server-store/internal/client"
	"github.com/trenchesdeveloper/mcp-server-store/internal/mcp"
)

// OrderToolSet groups all order-related tools and shares the HTTP client.
type OrderToolSet struct {
	httpClient *client.RestClient
	logger     *logrus.Logger
}

// NewOrderToolSet creates a new OrderToolSet with the given HTTP client and logger.
func NewOrderToolSet(httpClient *client.RestClient, logger *logrus.Logger) *OrderToolSet {
	return &OrderToolSet{httpClient: httpClient, logger: logger}
}

// ---- Create Order ----

// CreateOrderTool returns the tool definition for creating an order from the cart.
func (o *OrderToolSet) CreateOrderTool() mcp.Tool {
	return mcp.Tool{
		Name:        "place_order",
		Description: "Creates an order from the current shopping cart. Requires authentication.",
		InputSchema: mcp.InputSchema{
			Type: "object",
		},
	}
}

// CreateOrderHandler returns a handler that creates an order.
func (o *OrderToolSet) CreateOrderHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		o.logger.Info("Creating order from cart")

		body, err := o.httpClient.WithToken().Post("/orders", nil)
		if err != nil {
			o.logger.WithError(err).Error("Failed to create order")
			return nil, fmt.Errorf("failed to create order: %w", err)
		}

		var resp OrderResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			o.logger.WithError(err).Error("Failed to parse order response")
			return nil, fmt.Errorf("failed to parse order response: %w", err)
		}

		o.logger.WithFields(logrus.Fields{
			"order_id": resp.Data.ID,
			"total":    resp.Data.Total,
		}).Info("Order created")

		result := fmt.Sprintf("Order #%d created successfully!\n- Status: %s\n- Total: $%.2f",
			resp.Data.ID, resp.Data.Status, resp.Data.Total)

		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				mcp.NewTextContent(result),
			},
		}, nil
	}
}

// ---- List Orders ----

// ListOrdersTool returns the tool definition for listing orders.
func (o *OrderToolSet) ListOrdersTool() mcp.Tool {
	return mcp.Tool{
		Name:        "list_orders",
		Description: "Lists all orders for the current user with pagination. Requires authentication.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"page": {
					Type:        "string",
					Description: "Page number (default: 1)",
				},
				"limit": {
					Type:        "string",
					Description: "Items per page (default: 10)",
				},
			},
		},
	}
}

// ListOrdersHandler returns a handler that lists the user's orders.
func (o *OrderToolSet) ListOrdersHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		o.logger.Info("Listing orders")

		params := map[string]string{}
		for _, key := range []string{"page", "limit"} {
			if val, ok := arguments[key].(string); ok && val != "" {
				params[key] = val
			}
		}

		body, err := o.httpClient.WithToken().Get("/orders", params)
		if err != nil {
			o.logger.WithError(err).Error("Failed to list orders")
			return nil, fmt.Errorf("failed to list orders: %w", err)
		}

		var resp ListOrdersResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			o.logger.WithError(err).Error("Failed to parse orders response")
			return nil, fmt.Errorf("failed to parse orders response: %w", err)
		}

		o.logger.WithField("count", len(resp.Data)).Info("Orders listed")

		var sb strings.Builder
		fmt.Fprintf(&sb, "Found %d orders\n\n", len(resp.Data))

		for i, order := range resp.Data {
			fmt.Fprintf(&sb, "%d. Order #%d - %s - $%.2f\n",
				i+1, order.ID, order.Status, order.Total)
		}

		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				mcp.NewTextContent(sb.String()),
			},
		}, nil
	}
}

// ---- Cancel Order ----

// CancelOrderTool returns the tool definition for cancelling an order.
func (o *OrderToolSet) CancelOrderTool() mcp.Tool {
	return mcp.Tool{
		Name:        "cancel_order",
		Description: "Cancels a pending order. Only pending orders can be cancelled. Requires authentication.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"id": {
					Type:        "string",
					Description: "The order ID to cancel",
				},
			},
			Required: []string{"id"},
		},
	}
}

// CancelOrderHandler returns a handler that cancels an order.
func (o *OrderToolSet) CancelOrderHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		id, ok := arguments["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("order id is required")
		}

		o.logger.WithField("id", id).Info("Cancelling order")

		body, err := o.httpClient.WithToken().Post("/orders/"+id+"/cancel", nil)
		if err != nil {
			o.logger.WithError(err).Error("Failed to cancel order")
			return nil, fmt.Errorf("failed to cancel order: %w", err)
		}

		var resp OrderResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			o.logger.WithError(err).Error("Failed to parse cancel response")
			return nil, fmt.Errorf("failed to parse cancel response: %w", err)
		}

		o.logger.WithField("order_id", resp.Data.ID).Info("Order cancelled")

		result := fmt.Sprintf("Order #%d cancelled.\n- Status: %s\n- Total: $%.2f",
			resp.Data.ID, resp.Data.Status, resp.Data.Total)

		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				mcp.NewTextContent(result),
			},
		}, nil
	}
}
