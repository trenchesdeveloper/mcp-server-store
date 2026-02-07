package cart

import "time"

type CartItem struct {
	ID        uint      `json:"id"`
	Product struct {
		ID uint `json:"id"`
		Name string `json:"name"`
		Price float64 `json:"price"`
		Stock int `json:"stock"`
		Description string `json:"description"`
		Category struct {
			ID uint `json:"id"`
			Name string `json:"name"`
			Description string `json:"description"`
			IsActive bool `json:"is_active"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"category"`
	}
}

type CartResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    []CartItem `json:"data"`
	Error   string     `json:"error"`
}

type AddToCartRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type AddToCartResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    struct {
		ID uint `json:"id"`
		UserID uint `json:"user_id"`
		Total float64 `json:"total"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"data"`
	Error   string   `json:"error"`
}


type ViewCartResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    struct {
		ID uint `json:"id"`
		UserID uint `json:"user_id"`
		CartItems []CartItem `json:"cart_items"`
		Total float64 `json:"total"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	Error   string     `json:"error"`
}