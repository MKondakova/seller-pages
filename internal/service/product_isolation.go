package service

import (
	"context"
	"seller-pages-wb/internal/models"
	"sync"

	"go.uber.org/zap"
)

type PersonalProducts interface {
	GetProductsList(page int) ([]models.ProductPreview, int)
	GetProductByID(id string) (models.ProductPageInfo, error)
	AddProduct() models.ProductPreview
	DeleteProductByID(productID string) error
	GetProductsWithFeedbacks(page int) ([]models.FeedbackPageInfo, int)
}
type ProductIsolationService struct {
	services map[string]*ProductService

	initProducts     []models.Product
	feedbacksService *FeedbackService
	logger           *zap.SugaredLogger

	mu sync.RWMutex
}

func NewProductIsolationService(initProducts []models.Product, feedbackService *FeedbackService, logger *zap.SugaredLogger) *ProductIsolationService {
	return &ProductIsolationService{
		services:         make(map[string]*ProductService),
		initProducts:     initProducts,
		feedbacksService: feedbackService,
		logger:           logger,
		mu:               sync.RWMutex{},
	}
}

func (s *ProductIsolationService) GetProductsList(ctx context.Context, page int) ([]models.ProductPreview, int) {
	return s.getProductService(ctx).GetProductsList(page)
}
func (s *ProductIsolationService) GetProductByID(ctx context.Context, id string) (models.ProductPageInfo, error) {
	return s.getProductService(ctx).GetProductByID(id)
}
func (s *ProductIsolationService) AddProduct(ctx context.Context) models.ProductPreview {
	return s.getProductService(ctx).AddProduct()
}
func (s *ProductIsolationService) DeleteProductByID(ctx context.Context, productID string) error {
	return s.getProductService(ctx).DeleteProductByID(productID)
}
func (s *ProductIsolationService) GetProductsWithFeedbacks(ctx context.Context, page int) ([]models.FeedbackPageInfo, int) {
	return s.getProductService(ctx).GetProductsWithFeedbacks(page)
}

func (s *ProductIsolationService) getProductService(ctx context.Context) *ProductService {
	nickname := models.ClaimsFromContext(ctx).Nickname

	s.mu.RLock()
	service, has := s.services[nickname]
	s.mu.RUnlock()

	if has {
		return service
	}

	newService := NewProductService(s.initProducts, s.feedbacksService)

	s.mu.Lock()
	s.services[nickname] = newService
	s.mu.Unlock()

	s.logger.Infof("New Product isolation service with nickname %s created", nickname)

	return newService
}
