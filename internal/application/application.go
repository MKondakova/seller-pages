package application

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"seller-pages/internal/api"
	"seller-pages/internal/config"
	"seller-pages/internal/service"
	"seller-pages/pkg/runner"
)

type Application struct {
	cfg *config.Config

	productService  *service.ProductIsolationService
	balanceService  *service.BalanceService
	tokenService    *service.TokenService
	feedbackService *service.FeedbackService
	logger          *zap.SugaredLogger

	errChan chan error
	wg      sync.WaitGroup
	ready   bool
}

func New() *Application {
	return &Application{
		errChan: make(chan error),
	}
}

func (a *Application) Start(ctx context.Context) error {
	if err := a.initConfigAndLogger(); err != nil {
		return err
	}

	if err := a.initServices(); err != nil {
		return err
	}

	if err := a.initRouter(ctx); err != nil {
		return err
	}

	return nil
}

func (a *Application) Wait(ctx context.Context, cancel context.CancelFunc) error {
	var appErr error

	errWg := sync.WaitGroup{}

	errWg.Add(1)

	go func() {
		defer errWg.Done()

		for err := range a.errChan {
			cancel()
			a.logger.Error(err)
			appErr = err
		}
	}()

	<-ctx.Done()
	a.wg.Wait()
	close(a.errChan)
	errWg.Wait()

	return appErr
}

func (a *Application) Ready() bool {
	return a.ready
}

func (a *Application) HandleGracefulShutdown(ctx context.Context, cancel context.CancelFunc) error {
	var appErr error

	errWg := sync.WaitGroup{}

	errWg.Add(1)

	go func() {
		defer errWg.Done()

		for err := range a.errChan {
			cancel()
			a.logger.Error(err)
			appErr = err
		}
	}()

	<-ctx.Done()
	a.wg.Wait()
	close(a.errChan)
	errWg.Wait()

	return appErr
}

func (a *Application) initConfigAndLogger() error {
	if err := a.initLogger(); err != nil {
		return fmt.Errorf("can't init logger: %w", err)
	}

	if err := a.initConfig(); err != nil {
		return fmt.Errorf("can't init config: %w", err)
	}

	return nil
}

func (a *Application) initConfig() error {
	var err error

	a.cfg, err = config.GetConfig(a.logger)
	if err != nil {
		return fmt.Errorf("can't parse config: %w", err)
	}

	return nil
}

func (a *Application) initLogger() error {
	zapLog, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("can't create logger: %w", err)
	}

	a.logger = zapLog.Sugar()

	return nil
}

func (a *Application) initServices() error {
	var err error

	a.feedbackService, err = service.NewFeedbackService(
		"data/feedbacks.json",
		"data/feedbacksPerProduct.json",
		a.logger,
	)
	if err != nil {
		return fmt.Errorf("can't create feedback service: %w", err)
	}

	a.productService = service.NewProductIsolationService(a.cfg.InitialProductsData, a.feedbackService, a.logger)

	a.balanceService = service.NewBalanceService()
	a.tokenService = service.NewTokenService(a.cfg.PrivateKey, a.cfg.CreatedTokensPath)

	return nil
}

func (a *Application) initRouter(ctx context.Context) error {
	authMiddleware := api.NewAuthMiddleware(a.cfg.PublicKey, a.logger, a.cfg.RevokedTokens).JWTAuth

	router := api.NewRouter(
		a.cfg.ServerOpts,
		a.productService,
		a.balanceService,
		a.tokenService,
		authMiddleware,
		a.logger,
	)

	if err := runner.RunServer(ctx, router, a.cfg.ListenPort, a.errChan, &a.wg); err != nil {
		return fmt.Errorf("can't run public router: %w", err)
	}

	return nil
}
