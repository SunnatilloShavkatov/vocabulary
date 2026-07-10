package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	gwhttp "vocabulary/backend/gateway/internal/http"
	"vocabulary/backend/gateway/internal/grpcclient"
	"vocabulary/backend/libs/shared/config"
	"vocabulary/backend/libs/shared/db"
	notificationservice "vocabulary/backend/modules/notification/service"
	usersrepository "vocabulary/backend/modules/users/repository"
	usersservice "vocabulary/backend/modules/users/service"
	"vocabulary/backend/modules/vocabulary/repository"
	"vocabulary/backend/modules/vocabulary/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := db.NewPool(context.Background(), cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	rdb, err := db.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	vocabRepo := vocabularyrepository.NewVocabularyPgxRepository(pool)
	cachedVocabRepo := vocabularyrepository.NewCachedVocabularyRepository(vocabRepo, rdb)
	vocabularySvc := vocabularyservice.NewVocabularyService(cfg, cachedVocabRepo)

	usersRepo := usersrepository.NewUsersPgxRepository(pool)
	usersSvc := usersservice.NewUsersService(usersRepo)
	notificationSvc := notificationservice.NewNotificationService(nil)
	authGRPCTarget := strings.TrimSpace(os.Getenv("AUTH_GRPC_TARGET"))
	authGRPCClient := grpcclient.NewAuthClient(authGRPCTarget)
	defer func() {
		if err := authGRPCClient.Close(); err != nil {
			log.Printf("error closing auth grpc client connection: %v", err)
		}
	}()
	router := gwhttp.NewGatewayRouter(cfg.JWT.Secret, cfg.CORSAllowedOrigins, cfg.APIToken, vocabularySvc, usersSvc, notificationSvc, authGRPCClient)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("gateway listening on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped: %v", err)
	}
}
