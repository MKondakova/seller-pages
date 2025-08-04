package service

import (
	"math"

	"seller-pages-wb/internal/models"
)

const ProductsPerPage = 20

type ProductService struct {
	products []models.Product
}

func NewProductService() *ProductService {
	return &ProductService{
		products: []models.Product{
			{
				ID:                "123",
				Name:              "test-product",
				Article:           "test-article",
				Category:          "clothes",
				Description:       "Lorem ipsum description",
				ImageURL:          "https://storage.yandexcloud.net/std-ext-005-04-image/kb1k50ukjoq3v2qki637m0fj1oj6cdkb.jpg",
				IsRemovable:       true,
				OldPrice:          1500,
				Price:             100,
				Rating:            5,
				WarehouseQuantity: 100,
			},
		},
	}
}

func (s *ProductService) GetProductsList(page int) ([]models.ProductPreview, int) {
	totalPages := int(math.Ceil(float64(len(s.products)) / float64(ProductsPerPage)))

	paginationStart := (page - 1) * ProductsPerPage
	if paginationStart >= len(s.products) {
		return nil, totalPages
	}

	paginationEnd := paginationStart + ProductsPerPage
	if paginationEnd > len(s.products) {
		paginationEnd = len(s.products)
	}

	result := make([]models.ProductPreview, paginationEnd-paginationStart)

	for i, product := range s.products[paginationStart:paginationEnd] {
		result[i] = product.ToPreview()
	}

	return result, totalPages
}
