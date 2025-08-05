package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"

	"seller-pages-wb/internal/models"
)

const ProductsPerPage = 20

type ProductService struct {
	products     []models.Product
	productIndex map[string]*models.Product
}

func NewProductService(productsPath string) (*ProductService, error) {
	file, err := os.Open(productsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var products []models.Product
	if err := json.Unmarshal(bytes, &products); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	index := make(map[string]*models.Product, len(products))
	for i := range products {
		index[products[i].ID] = &products[i]
	}

	return &ProductService{
		products:     products,
		productIndex: index,
	}, nil
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

func (s *ProductService) GetProductsByID(productID string) (models.ProductPageInfo, error) {
	product, ok := s.productIndex[productID]
	if !ok {
		return models.ProductPageInfo{}, fmt.Errorf("%w: product %s not found", models.ErrNotFound, productID)
	}

	return product.ToPageInfo(), nil
}
