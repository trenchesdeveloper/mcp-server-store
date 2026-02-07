package orders

type Order struct {
	ID     uint    `json:"id"`
	Status string  `json:"status"`
	Total  float64 `json:"total"`
}

type OrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    Order  `json:"data"`
	Error   string `json:"error"`
}

type ListOrdersResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Data    []Order `json:"data"`
	Error   string  `json:"error"`
}
