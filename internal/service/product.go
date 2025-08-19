package service

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"

	"github.com/google/uuid"

	"seller-pages/internal/models"
)

const (
	ProductsPerPage = 20
	Tech            = "Электроника"
	Beauty          = "Косметика"
	Children        = "Детские товары"
	Clothes         = "Одежда"
	Household       = "Для дома"
	Stationery      = "Канцелярия"
)

var errProductLoss = errors.New("product loss")

var Categories = []string{
	Tech,
	Beauty,
	Children,
	Clothes,
	Household,
	Stationery,
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
	category := randomCategory()

	newProduct := models.Product{
		ID:                uuid.NewString(),
		Name:              randomName(category),
		Article:           randomArticle(),
		Category:          category,
		Description:       randomDescription(),
		ImageURL:          randomImageURL(category),
		IsRemovable:       rand.Float64() < 0.9,
		Rating:            randomRating(),
		WarehouseQuantity: randomWarehouseQuantity(),
		OrdersCount:       rand.Intn(1000),
		RefundsPercent:    rand.Float64() * 100,
	}

	newProduct.Price, newProduct.OldPrice = getRandomPriceAndOldPrice(category)

	s.productMutex.Lock()

	s.productIndex[newProduct.ID] = &newProduct
	s.products = append(s.products, newProduct)

	s.productMutex.Unlock()

	s.feedbackService.AddFeedbacksToProduct(newProduct)

	return newProduct.ToPreview()
}

func randomName(category string) string {
	var names []string
	switch category {
	case Household:
		names = []string{
			"Лампа настольная",
			"Стол рабочий",
			"Подставка для книг",
			"Декорации на стол",
			"Коробки для хранения",
		}
	case Tech:
		names = []string{
			"Монитор для игр",
			"Монитор FullHD",
			"Клавиатура геймерская с подсветкой",
			"Ноутбук",
			"Колонки стационарные",
		}
	case Beauty:
		names = []string{
			"Крем увлажняющий",
			"Крем омолаживающий",
			"Набор кремов",
			"Крем с фруктовым ароматом",
			"Крем для рук",
			"Крем для тела",
		}
	case Children:
		names = []string{
			"Комбнезон",
			"Одежда для выписки",
			"Белье детское",
			"Красивый комплект",
		}
	case Clothes:
		names = []string{
			"Футболка новой коллекции",
			"Классная кофта",
			"Идельные штаны",
			"Комплект на каждый день",
			"Комфортный комплект одежды",
		}
	case Stationery:
		names = []string{
			"Концелярский набор для школы",
			"Набор первоклассника",
			"Необходимые товары для офиса",
			"Ручки и карандаши с пеналом",
		}
	default:
		names = []string{
			"Товар что надо",
		}
	}

	return names[rand.Intn(len(names))]
}

func (s *ProductService) DeleteProductByID(productID string) error {
	s.productMutex.Lock()
	defer s.productMutex.Unlock()

	product, has := s.productIndex[productID]
	if !has {
		return fmt.Errorf("%w: product %s not found", models.ErrNotFound, productID)
	}

	if !product.IsRemovable {
		return fmt.Errorf("%w: product is not removable", models.ErrForbidden)
	}

	s.feedbackService.DeleteFeedbacks(productID)

	delete(s.productIndex, productID)
	for i, product := range s.products {
		if product.ID == productID {
			s.products = append(s.products[:i], s.products[i+1:]...)

			return nil
		}
	}

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

func randomImageURL(category string) string {
	var images []string
	switch category {
	case Household:
		images = []string{
			"https://basket-15.wbbasket.ru/vol2271/part227133/227133697/images/big/8.webp",
			"https://basket-15.wbbasket.ru/vol2271/part227133/227133697/images/big/9.webp",
			"https://basket-15.wbbasket.ru/vol2271/part227133/227133697/images/big/10.webp",
			"https://basket-15.wbbasket.ru/vol2271/part227133/227133697/images/big/10.webp",
		}
	case Tech:
		images = []string{
			"https://basket-15.wbbasket.ru/vol2405/part240554/240554144/images/big/3.webp",
			"https://basket-15.wbbasket.ru/vol2405/part240554/240554144/images/big/4.webp",
			"https://basket-26.wbbasket.ru/vol4665/part466522/466522013/images/big/1.webp",
		}
	case Beauty:
		images = []string{
			"https://basket-10.wbbasket.ru/vol1511/part151190/151190621/images/big/1.webp",
			"https://basket-11.wbbasket.ru/vol1625/part162546/162546677/images/big/2.webp",
			"https://basket-19.wbbasket.ru/vol3178/part317854/317854683/images/big/1.webp",
			"https://basket-19.wbbasket.ru/vol3178/part317854/317854683/images/big/3.webp",
			"https://basket-16.wbbasket.ru/vol2484/part248439/248439960/images/big/1.webp",
			"https://basket-17.wbbasket.ru/vol2679/part267903/267903164/images/big/1.webp",
		}
	case Children:
		images = []string{
			"https://basket-17.wbbasket.ru/vol2699/part269922/269922591/images/big/2.webp",
			"https://basket-19.wbbasket.ru/vol3153/part315358/315358478/images/big/6.webp",
			"https://basket-18.wbbasket.ru/vol3047/part304775/304775028/images/big/2.webp",
			"https://basket-18.wbbasket.ru/vol3047/part304775/304775028/images/big/3.webp",
			"https://basket-26.wbbasket.ru/vol4777/part477756/477756673/images/big/1.webp",
		}
	case Clothes:
		images = []string{
			"https://basket-02.wbbasket.ru/vol255/part25539/25539349/images/big/2.webp",
			"https://basket-10.wbbasket.ru/vol1314/part131439/131439248/images/big/3.webp",
			"https://basket-15.wbbasket.ru/vol2237/part223714/223714031/images/big/1.webp",
			"https://basket-05.wbbasket.ru/vol963/part96316/96316155/images/big/3.webp",
			"https://basket-26.wbbasket.ru/vol4806/part480669/480669352/images/big/2.webp",
		}
	case Stationery:
		images = []string{
			"https://basket-16.wbbasket.ru/vol2581/part258124/258124707/images/big/1.webp",
			"https://basket-12.wbbasket.ru/vol1712/part171222/171222754/images/big/1.webp",
			"https://basket-21.wbbasket.ru/vol3533/part353384/353384700/images/big/1.webp",
			"https://basket-25.wbbasket.ru/vol4458/part445898/445898947/images/big/1.webp",
			"https://basket-16.wbbasket.ru/vol2501/part250150/250150130/images/big/1.webp",
			"https://basket-26.wbbasket.ru/vol4583/part458372/458372626/images/big/1.webp",
		}
	default:
		images = []string{
			"https://basket-16.wbbasket.ru/vol2574/part257400/257400077/images/big/11.webp",
			"https://basket-22.wbbasket.ru/vol3704/part370494/370494201/images/big/2.webp",
			"https://basket-16.wbbasket.ru/vol2619/part261968/261968456/images/big/1.webp",
			"https://basket-09.wbbasket.ru/vol1207/part120753/120753915/images/big/1.webp",
		}
	}

	return images[rand.Intn(len(images))]
}

func getRandomPriceAndOldPrice(category string) (price, oldPrice float64) {
	var minPrice, maxPrice float64
	switch category {
	case Household:
		fallthrough
	case Tech:
		minPrice = 5000
		maxPrice = 20000
	case Beauty:
		minPrice = 50
		maxPrice = 1000
	case Children:
		minPrice = 100
		maxPrice = 10000
	case Clothes:
		minPrice = 1000
		maxPrice = 10000
	case Stationery:
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
