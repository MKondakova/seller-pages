package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"seller-pages/internal/config"
	"seller-pages/internal/models"

	"github.com/rs/cors"
	"go.uber.org/zap"
)

var (
	errInvalidPageNumber = errors.New("invalid page number")
	errEmptyID           = errors.New("empty id")
	errEmptyName         = errors.New("empty name")
)

type ProductsService interface {
	GetProductsList(ctx context.Context, page int) ([]models.ProductPreview, int)
	GetProductByID(ctx context.Context, id string) (models.ProductPageInfo, error)
	AddProduct(ctx context.Context) models.ProductPreview
	DeleteProductByID(ctx context.Context, productID string) error
	GetProductsWithFeedbacks(ctx context.Context, page int) ([]models.FeedbackPageInfo, int)
}

type BalanceService interface {
	GetBalanceInfo() models.BalanceInfo
}

type TokenService interface {
	GenerateToken(ctx context.Context, username string, isTeacher bool) (string, error)
}

type Router struct {
	*http.Server
	router *http.ServeMux

	productsService ProductsService
	balanceService  BalanceService
	tokenService    TokenService

	logger *zap.SugaredLogger
}

func NewRouter(
	cfg config.ServerOpts,
	productsService ProductsService,
	balanceService BalanceService,
	tokenService TokenService,
	authMiddleware func(next http.HandlerFunc) http.HandlerFunc,
	logger *zap.SugaredLogger,
) *Router {
	innerRouter := http.NewServeMux()

	appRouter := &Router{
		Server: &http.Server{
			Handler:      cors.AllowAll().Handler(innerRouter),
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		},
		router:          innerRouter,
		productsService: productsService,
		balanceService:  balanceService,
		tokenService:    tokenService,
		logger:          logger,
	}

	innerRouter.HandleFunc("POST /api/products/generate", authMiddleware(appRouter.addProduct))
	innerRouter.HandleFunc("GET /api/products", authMiddleware(appRouter.getProductsList))

	innerRouter.HandleFunc("GET /api/products/{id}", authMiddleware(appRouter.getProductByID))
	innerRouter.HandleFunc("DELETE /api/products/{id}", authMiddleware(appRouter.deleteProductByID))

	innerRouter.HandleFunc("GET /api/balanceInfo", authMiddleware(appRouter.getBalanceInfo))

	innerRouter.HandleFunc("POST /api/createToken", authMiddleware(appRouter.createToken))
	innerRouter.HandleFunc("POST /api/createTeacherToken", authMiddleware(appRouter.createTeacherToken))
	innerRouter.HandleFunc("GET /api/feedbacks", authMiddleware(appRouter.getFeedbacks))
	innerRouter.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "redoc-static.html")
	})

	return appRouter
}

func (r *Router) sendResponse(response http.ResponseWriter, request *http.Request, code int, buf []byte) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(code)
	_, err := response.Write(buf)
	if err != nil {
		r.logger.With(
			"module", "api",
			"request_url", request.Method+": "+request.URL.Path,
		).Errorf("Error sending error response: %v", err)
	}
}

func (r *Router) sendErrorResponse(response http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, models.ErrBadRequest):
		response.WriteHeader(http.StatusBadRequest)
		r.logger.With(
			"module", "api",
			"request_url", request.Method+": "+request.URL.Path,
		).Warn(err)
		r.writeError(response, request, err)

		return
	case errors.Is(err, models.ErrNotFound):
		response.WriteHeader(http.StatusNotFound)
		r.logger.With(
			"module", "api",
			"request_url", request.Method+": "+request.URL.Path,
		).Warn(err)

		r.writeError(response, request, err)

		return
	case errors.Is(err, models.ErrForbidden):
		response.WriteHeader(http.StatusForbidden)
		r.logger.With(
			"module", "api",
			"request_url", request.Method+": "+request.URL.Path,
		).Warn(err)

		r.writeError(response, request, err)

		return
	case errors.Is(err, models.ErrUnauthorized):
		response.WriteHeader(http.StatusUnauthorized)
		r.logger.With(
			"module", "api",
			"request_url", request.Method+": "+request.URL.Path,
		).Warn(err)

		r.writeError(response, request, err)

		return
	}

	response.WriteHeader(http.StatusInternalServerError)
	r.logger.With(
		"module", "api",
		"request_url", request.Method+": "+request.URL.Path,
	).Error(err)

	r.writeError(response, request, err)
}

func (r *Router) writeError(response http.ResponseWriter, request *http.Request, err error) {
	_, err = response.Write([]byte(err.Error()))
	if err != nil {
		r.logger.With(
			"module", "api",
			"request_url", request.Method+": "+request.URL.Path,
		).Errorf("Error sending error response: %v", err)
	}
}

func (r *Router) getProductsList(writer http.ResponseWriter, request *http.Request) {
	page, err := getPage(request)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrBadRequest, err))

		return
	}

	result, totalPages := r.productsService.GetProductsList(request.Context(), page)

	responseBody := PaginatedResponse[models.ProductPreview]{
		TotalPages: totalPages,
		Data:       result,
		Page:       page,
	}

	buf, err := json.Marshal(responseBody)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrInternalServer, err))

		return
	}

	r.sendResponse(writer, request, http.StatusOK, buf)
}

func (r *Router) getProductByID(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if id == "" {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrBadRequest, errEmptyID))

		return
	}

	product, err := r.productsService.GetProductByID(request.Context(), id)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("GetProductByID: %w", err))

		return
	}

	buf, err := json.Marshal(product)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrInternalServer, err))

		return
	}

	r.sendResponse(writer, request, http.StatusOK, buf)
}

func (r *Router) getBalanceInfo(writer http.ResponseWriter, request *http.Request) {
	responseBody := r.balanceService.GetBalanceInfo()

	buf, err := json.Marshal(responseBody)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrInternalServer, err))

		return
	}

	r.sendResponse(writer, request, http.StatusOK, buf)

}

func (r *Router) addProduct(writer http.ResponseWriter, request *http.Request) {
	responseBody := r.productsService.AddProduct(request.Context())

	buf, err := json.Marshal(responseBody)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrInternalServer, err))

		return
	}

	r.sendResponse(writer, request, http.StatusOK, buf)
}

func (r *Router) deleteProductByID(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if id == "" {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrBadRequest, errEmptyID))

		return
	}

	err := r.productsService.DeleteProductByID(request.Context(), id)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("GetProductByID: %w", err))

		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (r *Router) createToken(writer http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("name")
	if name == "" {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrBadRequest, errEmptyName))

		return
	}

	token, err := r.tokenService.GenerateToken(request.Context(), name, false)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("CreateToken: %w", err))

		return
	}

	r.sendResponse(writer, request, http.StatusOK, []byte(token))
}

func (r *Router) createTeacherToken(writer http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("name")
	if name == "" {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrBadRequest, errEmptyName))

		return
	}

	token, err := r.tokenService.GenerateToken(request.Context(), name, true)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("CreateToken: %w", err))

		return
	}

	responseBody := TokenResponse{
		Token: token,
	}

	buf, err := json.Marshal(responseBody)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrInternalServer, err))

		return
	}

	r.sendResponse(writer, request, http.StatusOK, buf)
}

func (r *Router) getFeedbacks(writer http.ResponseWriter, request *http.Request) {
	page, err := getPage(request)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrBadRequest, err))

		return
	}

	result, totalPages := r.productsService.GetProductsWithFeedbacks(request.Context(), page)

	responseBody := PaginatedResponse[models.FeedbackPageInfo]{
		TotalPages: totalPages,
		Data:       result,
		Page:       page,
	}

	buf, err := json.Marshal(responseBody)
	if err != nil {
		r.sendErrorResponse(writer, request, fmt.Errorf("%w: %w", models.ErrInternalServer, err))

		return
	}

	r.sendResponse(writer, request, http.StatusOK, buf)
}

func getPage(request *http.Request) (int, error) {
	pageParameter := request.URL.Query().Get("page")

	if pageParameter == "" {
		return 1, nil
	}

	page, err := strconv.Atoi(pageParameter)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", errInvalidPageNumber, err)
	}

	if page <= 0 {
		return 0, fmt.Errorf("%w: %d", errInvalidPageNumber, page)
	}

	return page, nil
}
