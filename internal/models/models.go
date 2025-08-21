package models

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type Product struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Article           string  `json:"article"`
	Category          string  `json:"category"`
	Description       string  `json:"description"`
	ImageURL          string  `json:"imageUrl"`
	IsRemovable       bool    `json:"isRemovable"`
	OldPrice          float64 `json:"oldPrice,omitempty"`
	Price             float64 `json:"price"`
	Rating            float64 `json:"rating,omitempty"`
	WarehouseQuantity int     `json:"warehouseQuantity,omitempty"`
	OrdersCount       int     `json:"ordersCount,omitempty"`
	RefundsPercent    float64 `json:"refundsPercent,omitempty"`
}
type ProductPageInfo struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Article           string  `json:"article"`
	Category          string  `json:"category"`
	Description       string  `json:"description"`
	ImageURL          string  `json:"imageUrl"`
	OldPrice          float64 `json:"oldPrice,omitempty"`
	Price             float64 `json:"price"`
	Rating            float64 `json:"rating,omitempty"`
	WarehouseQuantity int     `json:"warehouseQuantity,omitempty"`
	OrdersCount       int     `json:"ordersCount,omitempty"`
}

type ProductPreview struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Article     string  `json:"article"`
	Category    string  `json:"category"`
	ImageURL    string  `json:"imageUrl"`
	IsRemovable bool    `json:"isRemovable"`
	OldPrice    float64 `json:"oldPrice,omitempty"`
	Price       float64 `json:"price"`
}

func (p *Product) ToPreview() ProductPreview {
	return ProductPreview{
		ID:          p.ID,
		Name:        p.Name,
		Article:     p.Article,
		Category:    p.Category,
		ImageURL:    p.ImageURL,
		IsRemovable: p.IsRemovable,
		OldPrice:    p.OldPrice,
		Price:       p.Price,
	}
}

func (p *Product) ToPageInfo() ProductPageInfo {
	return ProductPageInfo{
		ID:                p.ID,
		Name:              p.Name,
		Article:           p.Article,
		Category:          p.Category,
		Description:       p.Description,
		ImageURL:          p.ImageURL,
		OldPrice:          p.OldPrice,
		Price:             p.Price,
		Rating:            p.Rating,
		WarehouseQuantity: p.WarehouseQuantity,
		OrdersCount:       p.OrdersCount,
	}
}

type BalanceInfo struct {
	ShopID              string       `json:"shopId"`
	Balance             float64      `json:"balance"`
	Sales               float64      `json:"sales"`
	Income              float64      `json:"income"`
	ShopRating          float32      `json:"shopRating"`
	TotalSalesCount     int          `json:"totalSalesCount"`
	TotalRefundsCount   int          `json:"totalRefundsCount"`
	SalesChart          SaleChartDTO `json:"salesChart"`
	TotalRefundsPercent float32      `json:"totalRefundsPercent"`
	MonthlyRatingGrow   float32      `json:"monthlyRatingGrow,omitempty"`
	MonthlySalesGrow    int          `json:"monthlySalesGrow,omitempty"`
}

type SaleChartDTO struct {
	AverageSales float64     `json:"averageSales"`
	Data         []SalePoint `json:"data"`
}

type SalePoint struct {
	Amount float64 `json:"amount"`
	Period string  `json:"period"`
}

type FeedbackPageInfo struct {
	ID             string      `json:"id"`
	ImageURL       string      `json:"imageUrl"`
	Name           string      `json:"name"`
	Rating         float64     `json:"rating,omitempty"`
	OrdersCount    int         `json:"ordersCount"`
	RefundsPercent float64     `json:"refundsPercent,omitempty"`
	Feedbacks      []*Feedback `json:"feedbacks"`
}
type Feedback struct {
	ID        string   `json:"id"`
	BuyerName string   `json:"buyerName"`
	Rating    int      `json:"rating"`
	Pros      string   `json:"pros"`
	Cons      string   `json:"cons"`
	Comment   string   `json:"comment"`
	PhotosURL []string `json:"photosURL"`
	IsRefund  bool     `json:"isRefund"`
}

type AuthTokenClaims struct {
	*jwt.RegisteredClaims

	Nickname  string `json:"nickname"`
	IsTeacher bool   `json:"isTeacher"`
}

type ContextClaimsKey struct{}

func ClaimsFromContext(ctx context.Context) *AuthTokenClaims {
	claims, _ := ctx.Value(ContextClaimsKey{}).(*AuthTokenClaims)

	return claims
}
