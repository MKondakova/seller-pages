package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"seller-pages-wb/internal/models"
)

type FeedbackService struct {
	feedbacks           map[string]*models.Feedback
	feedbacksPerProduct map[string][]string

	mx sync.RWMutex
}

func NewFeedbackService(feedbacksPath, feedbacksIndexPath string) (*FeedbackService, error) {
	result := &FeedbackService{
		feedbacks:           make(map[string]*models.Feedback),
		feedbacksPerProduct: make(map[string][]string),
	}

	file, err := os.Open(feedbacksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(bytes, &result.feedbacks); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	file, err = os.Open(feedbacksIndexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	bytes, err = io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(bytes, &result.feedbacksPerProduct); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

func (s *FeedbackService) GetFeedbacks(product models.Product) models.FeedbackPageInfo {
	s.mx.RLock()
	defer s.mx.RUnlock()

	feedbacks := s.feedbacksPerProduct[product.ID]

	result := models.FeedbackPageInfo{
		ID:             product.ID,
		Name:           product.Name,
		Rating:         product.Rating,
		OrdersCount:    product.OrdersCount,
		RefundsPercent: product.RefundsPercent,
		Feedbacks:      make([]*models.Feedback, len(feedbacks)),
	}

	for i, id := range feedbacks {
		result.Feedbacks[i] = s.feedbacks[id]
	}

	return result
}

func (s *FeedbackService) DeleteFeedbacks(productID string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	delete(s.feedbacksPerProduct, productID)
	delete(s.feedbacks, productID)
}

func (s *FeedbackService) AddFeedbacksToProduct(product models.Product) {
	s.mx.Lock()
	defer s.mx.Unlock()

	number := rand.Intn(5)

	feedbackIDs := make([]string, number)
	for i := range number {
		id := uuid.NewString()
		feedbackIDs[i] = id

		rating := rand.Intn(4) + 1
		s.feedbacks[id] = &models.Feedback{
			ID:        id,
			BuyerName: "Aлександр" + strconv.Itoa(rand.Intn(1000)),
			Rating:    rating,
			Pros:      getRandomPros(rating),
			Cons:      getRandomCons(rating),
			Comment:   getRandomComment(),
			PhotosURL: getRandomPhotosForFeedback(rand.Intn(4)),
			IsRefund:  rand.Float64() < 0.3,
		}
	}

	s.feedbacksPerProduct[product.ID] = feedbackIDs
}

func getRandomPros(rating int) string {
	if rand.Float64() < 0.3 {
		return ""
	}

	if rating >= 3 {
		return "Товар хороший, можно брать, я вот взял"
	}

	return "Их не много"
}

func getRandomCons(rating int) string {
	if rand.Float64() < 0.3 {
		return ""
	}

	if rating >= 3 {
		return "НУ в принципе их нет"
	}

	return "Никогда больше ничего у вас не куплю"
}

func getRandomComment() string {
	return "Скоро я закажу себе всю корзину вб и отзывы станут полнее, но пока вот так"
}

func getRandomPhotosForFeedback(n int) []string {
	result := make([]string, n)
	for i := range result {
		result[i] = randomImageURL()
	}

	return result
}
