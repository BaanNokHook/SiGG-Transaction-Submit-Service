package app

import (
	"fmt"
	"nextclan/transaction-gateway/transaction-submit-service/config"
	v1 "nextclan/transaction-gateway/transaction-submit-service/internal/controller/http/v1"
	usecase "nextclan/transaction-gateway/transaction-submit-service/internal/usecase/transaction"
	"nextclan/transaction-gateway/transaction-submit-service/pkg/httpserver"
	"nextclan/transaction-gateway/transaction-submit-service/pkg/loaffinity"
	"nextclan/transaction-gateway/transaction-submit-service/pkg/logger"
	messaging "nextclan/transaction-gateway/transaction-submit-service/pkg/rabbitmq"
	"nextclan/transaction-gateway/transaction-submit-service/pkg/redis"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

type sampleMessage struct {
}

func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)
	fmt.Println("Starting App...")

	// Init redis
	redisCache := redis.NewRedisClient(redis.Addr(cfg.Redis.Addr), redis.Password(cfg.Redis.Password))

	// Use case
	receiveValidatedTransactionUseCase := usecase.NewReceiveValidatedTransaction(l, redisCache)

	// HTTP Server
	httpServer := initializeHttp(l, cfg)

	//Init client
	initializeMessaging(cfg, receiveValidatedTransactionUseCase)
	initializeLoaffinityRPC(cfg, l)

	// Shutdown
	ShutdownApplicationHandler(l, httpServer)
}

func initializeHttp(l *logger.Logger, cfg *config.Config) *httpserver.Server {
	handler := gin.New()
	v1.NewRouter(handler, l)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))
	return httpServer
}

func initializeMessaging(cfg *config.Config, vt *usecase.ReceiveValidatedTransactionFromQueueUseCase) {
	//TODO dependency injection for usecase scope
	usecase.MessagingClient = &messaging.MessagingClient{}
	usecase.MessagingClient.Connect(cfg.RMQ.URL)
	usecase.MessagingClient.SubscribeToQueue("txt.gw", "topic", "validated.transaction", "transaction.submit.service", vt.Handle)
}

func initializeLoaffinityRPC(cfg *config.Config, l *logger.Logger) {
	usecase.LoaffinityClient = loaffinity.NewLoaffinityClient(cfg.Loaffinity.URL, cfg.Loaffinity.Username, cfg.Loaffinity.Password, l)
}

func ShutdownApplicationHandler(l *logger.Logger, httpServer *httpserver.Server) {
	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	}

	err := httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

	err = usecase.MessagingClient.Close()
}
