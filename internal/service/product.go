package service

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"

	"github.com/google/uuid"

	"seller-pages-wb/internal/models"
)

const ProductsPerPage = 20

var errProductLoss = errors.New("product loss")

var Categories = []string{
	"Электроника",
	"Косметика",
	"Детские тоdefer s.productMutex.RUnlock()вары",
	"Одежда",
	"Бытовая техника",
	"Канцелярия",
}

type FeedbackProvider interface {
	GetFeedbacks(product models.Product) models.FeedbackPageInfo
	AddFeedbacksToProduct(product models.Product)
	DeleteFeedbacks(product string)
}

type ProductService struct {
	products        []models.Product
	productIndex    map[string]*models.Product
	feedbackService FeedbackProvider

	productMutex sync.RWMutex
}

func NewProductService(products []models.Product, feedbackService FeedbackProvider) *ProductService {
	index := make(map[string]*models.Product, len(products))
	for i := range products {
		index[products[i].ID] = &products[i]
	}

	return &ProductService{
		products:        products,
		productIndex:    index,
		feedbackService: feedbackService,
	}
}

func (s *ProductService) GetProductsList(page int) ([]models.ProductPreview, int) {
	s.productMutex.RLock()
	productsAmount := len(s.products)
	s.productMutex.RUnlock()

	totalPages := int(math.Ceil(float64(productsAmount) / float64(ProductsPerPage)))

	paginationStart := (page - 1) * ProductsPerPage
	if paginationStart >= productsAmount {
		return nil, totalPages
	}

	paginationEnd := paginationStart + ProductsPerPage
	if paginationEnd > productsAmount {
		paginationEnd = productsAmount
	}

	listLen := paginationEnd - paginationStart
	result := make([]models.ProductPreview, listLen)
	productsToTransform := make([]models.Product, listLen)

	s.productMutex.RLock()
	_ = copy(productsToTransform, s.products[paginationStart:paginationEnd])
	s.productMutex.RUnlock()

	for i, product := range productsToTransform {
		result[i] = product.ToPreview()
	}

	return result, totalPages
}

func (s *ProductService) GetProductsWithFeedbacks(page int) ([]models.FeedbackPageInfo, int) {
	s.productMutex.RLock()
	productsAmount := len(s.products)
	s.productMutex.RUnlock()

	totalPages := int(math.Ceil(float64(productsAmount) / float64(ProductsPerPage)))

	paginationStart := (page - 1) * ProductsPerPage
	if paginationStart >= productsAmount {
		return nil, totalPages
	}

	paginationEnd := paginationStart + ProductsPerPage
	if paginationEnd > productsAmount {
		paginationEnd = productsAmount
	}

	listLen := paginationEnd - paginationStart
	result := make([]models.FeedbackPageInfo, listLen)
	productsToTransform := make([]models.Product, listLen)

	s.productMutex.RLock()
	_ = copy(productsToTransform, s.products[paginationStart:paginationEnd])
	s.productMutex.RUnlock()

	for i, product := range productsToTransform {
		result[i] = s.feedbackService.GetFeedbacks(product)
	}

	return result, totalPages
}

func (s *ProductService) GetProductByID(productID string) (models.ProductPageInfo, error) {
	s.productMutex.RLock()
	defer s.productMutex.RUnlock()

	product, ok := s.productIndex[productID]

	if !ok {
		return models.ProductPageInfo{}, fmt.Errorf("%w: product %s not found", models.ErrNotFound, productID)
	}

	return product.ToPageInfo(), nil
}

func (s *ProductService) AddProduct() models.ProductPreview {
	newProduct := models.Product{
		ID:                uuid.NewString(),
		Name:              randomName(),
		Article:           randomArticle(),
		Category:          randomCategory(),
		Description:       randomDescription(),
		ImageURL:          randomImageURL(),
		IsRemovable:       rand.Float64() < 0.7,
		Rating:            randomRating(),
		WarehouseQuantity: randomWarehouseQuantity(),
		OrdersCount:       rand.Intn(1000),
		RefundsPercent:    rand.Float64() * 100,
	}

	newProduct.Price, newProduct.OldPrice = getRandomPriceAndOldPrice(newProduct.Category)

	s.productMutex.Lock()

	s.productIndex[newProduct.ID] = &newProduct
	s.products = append(s.products, newProduct)

	s.productMutex.Unlock()

	s.feedbackService.AddFeedbacksToProduct(newProduct)

	return newProduct.ToPreview()
}

func randomName() string {
	return "Временное имя " + strconv.Itoa(rand.Intn(100))
}

func (s *ProductService) DeleteProductByID(productID string) error {
	_, has := s.productIndex[productID]
	if !has {
		return fmt.Errorf("%w: product %s not found", models.ErrNotFound, productID)
	}

	s.productMutex.Lock()
	defer s.productMutex.Unlock()

	delete(s.productIndex, productID)
	for i, product := range s.products {
		if product.ID == productID {
			s.products = append(s.products[:i], s.products[i+1:]...)

			return nil
		}
	}

	s.feedbackService.DeleteFeedbacks(productID)

	return fmt.Errorf("%w: product not found in list", errProductLoss)
}

func randomWarehouseQuantity() int {
	return rand.Intn(1000)
}

func randomRating() float64 {
	return rand.Float64() * 5
}

func randomCategory() string {
	return Categories[rand.Intn(len(Categories))]
}

func randomArticle() string {
	articleMin := 1000000000
	articleMax := 9999999999

	article := rand.Intn(articleMax-articleMin) + articleMin

	return strconv.Itoa(article)
}

func randomDescription() string {
	descriptions := []string{
		"Отличный выбор для повседневного использования. Подходит для всех возрастов и прост в эксплуатации.",
		"Идеально подходит как для дома, так и для офиса. Стильный дизайн и практичность делают этот товар незаменимым.",
		"Этот товар будет служить вам долго, его внешний вид не изменится со временем независимо от условий эксплуатации.",
		"Наверняка, вам не хватало именно этого товара в вашей коллекции. Став обладателем данной позиции вы забудете что такое скука.",
	}

	description := descriptions[rand.Intn(len(descriptions))] + " Берите, не пожалеете!🐈"

	return description
}

func randomImageURL() string {
	images := []string{
		"https://basket-16.wbbasket.ru/vol2574/part257400/257400077/images/big/11.webp",
		"https://basket-22.wbbasket.ru/vol3704/part370494/370494201/images/big/2.webp",
		"https://basket-16.wbbasket.ru/vol2619/part261968/261968456/images/big/1.webp",
		"https://basket-09.wbbasket.ru/vol1207/part120753/120753915/images/big/1.webp",
	}

	return images[rand.Intn(len(images))]
}

func getRandomPriceAndOldPrice(category string) (price, oldPrice float64) {
	var minPrice, maxPrice float64
	switch category {
	case Categories[4]:
		fallthrough
	case Categories[0]:
		minPrice = 5000
		maxPrice = 20000
	case Categories[1]:
		minPrice = 50
		maxPrice = 1000
	case Categories[2]:
		minPrice = 100
		maxPrice = 10000
	case Categories[3]:
		minPrice = 1000
		maxPrice = 10000
	case Categories[5]:
		minPrice = 10
		maxPrice = 500
	}

	price = rand.Float64()*(maxPrice-minPrice) + minPrice

	const probabilityOfEmptyOldPrice = 0.7
	if rand.Float64() < probabilityOfEmptyOldPrice {
		oldPrice = price + rand.Float64()*(maxPrice-price)
	}

	return
}
