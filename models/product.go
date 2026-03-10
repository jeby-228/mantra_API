package models

// Product represents a product stored in PostgreSQL and managed by GORM.
type Product struct {
	ProductName        string  `gorm:"size:255;not null" json:"product_name"`
	ProductPrice       float64 `gorm:"not null"          json:"product_price"`
	ProductDescription string  `gorm:"size:255"          json:"product_description"`
	ProductImage       string  `gorm:"size:255"          json:"product_image"`
	ProductStock       int     `gorm:"not null"          json:"product_stock"`
	Base
}
