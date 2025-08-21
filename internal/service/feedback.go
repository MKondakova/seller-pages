package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"seller-pages/internal/models"
)

type FeedbackService struct {
	feedbacks           map[string]*models.Feedback
	feedbacksPerProduct map[string][]string

	logger *zap.SugaredLogger
	mx     sync.RWMutex
}

func NewFeedbackService(feedbacksPath, feedbacksIndexPath string, logger *zap.SugaredLogger) (*FeedbackService, error) {
	result := &FeedbackService{
		feedbacks:           make(map[string]*models.Feedback),
		feedbacksPerProduct: make(map[string][]string),
		logger:              logger,
		mx:                  sync.RWMutex{},
	}

	file, err := os.Open(feedbacksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Errorf("failed to close feedbacks file: %v", err)
		}
	}(file)

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
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Errorf("failed to close feedbacks index file: %v", err)
		}
	}(file)

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
		ImageURL:       product.ImageURL,
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
			BuyerName: getRandomName(),
			Rating:    rating,
			Pros:      getRandomPros(rating),
			Cons:      getRandomCons(rating),
			Comment:   getRandomComment(),
			PhotosURL: getRandomPhotosForFeedback(rand.Intn(4), product.Category),
			IsRefund:  rand.Float64() < 0.3,
		}
	}

	s.feedbacksPerProduct[product.ID] = feedbackIDs
}

func getRandomName() string {
	names := []string{
		"Анна Петрова",
		"Сергей Иванов",
		"Мария Кузнецова",
		"Ирина Смирнова",
		"Алексей Волков",
		"Дмитрий Соколов",
		"Ольга Морозова",
		"Николай Федоров",
		"Елена Васильева",
		"Андрей Николаев",
		"Виктория Павлова",
		"Татьяна Мельникова",
		"Михаил Крылов",
		"John D.",
		"Emily R.",
		"Лена",
		"Вера",
	}

	return names[rand.Intn(len(names))]
}

func getRandomPros(rating int) string {
	if rand.Float64() < 0.3 {
		return ""
	}

	if rating >= 3 {
		pros := []string{
			"Товар хороший, можно брать, я вот взял",
			"Всё пришло целое, без повреждений",
			"Идеально подошло, радости моей нет предела",
			"Упаковка была аккуратной и плотной",
			"Работает стабильно, без перебоев",
			"Выглядит аккуратно и без лишних деталей",
			"Цена показалась адекватной за полученное качество",
		}

		return pros[rand.Intn(len(pros))]
	}

	return "Их не много"
}

func getRandomCons(rating int) string {
	if rand.Float64() < 0.3 || rating == 5 {
		return ""
	}

	if rating >= 3 {
		cons := []string{
			"Доставка заняла больше времени, чем хотелось бы",
			"Упаковка могла быть надёжнее",
			"Цена кажется слегка завышенной",
			"Некоторые детали показались не очень качественными",
			"Есть ощущение, что прослужит недолго",
			"При получении заметил(а) мелкие недочёты",
		}

		return cons[rand.Intn(len(cons))]
	}

	negativeCons := []string{
		"Никогда больше ничего у вас не куплю",
		"Ужасный товар, сплошное разочарование",
		"О, их так много, что и не описать",
	}

	return negativeCons[rand.Intn(len(negativeCons))]
}

func getRandomComment() string {
	comments := []string{
		"Доставка даже опередила ожидания, супер быстро, даже очереди в пункте выдачи не было. ",
		"Много времени заняло сравнение с конкурентами, супер много аналогов, но выбор пал на этот. ",
		"По сроку службы сложно сказать, попользоваться надо, отзыв будет дополняться. ",
		"Товар как товар, ожидания совпали с реальностью, на фото все видно. ",
		"Пользуюсь недолго, пока никаких особых впечатлений нет. По качеству всё в пределах нормы. ",
		"Пока не понятно, насколько удобно, нужно больше времени. Обычная вещь, без лишних деталей. ",
	}

	firstPart := rand.Intn(len(comments))
	comment := comments[firstPart]
	if secondPart := rand.Intn(len(comments)); rand.Float64() < 0.3 && secondPart != firstPart {
		comment += comments[secondPart]
	}

	return comment
}

func getRandomPhotosForFeedback(n int, category string) []string {
	result := make([]string, n)
	for i := range result {
		result[i] = randomImageURL(category)
	}

	return result
}
