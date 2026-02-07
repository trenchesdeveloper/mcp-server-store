package products

import "time"

type Product struct {
	ID          uint           `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	Stock       int            `json:"stock"`
	CategoryID  uint           `json:"category_id"`
	SKU         string         `json:"sku"`
	IsActive    bool           `json:"is_active"`
	Category    Category       `json:"category"`
	Images      []ProductImage `json:"images"`
}

type ProductImage struct {
	ID        uint      `json:"id"`
	URL       string    `json:"url"`
	AltText   string    `json:"alt_text"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type ProductResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    []Product `json:"data"`
	Meta    Meta      `json:"meta"`
	Error   string    `json:"error"`
}

type Meta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

type ProductDetailResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Data    Product `json:"data"`
	Error   string  `json:"error"`
}
