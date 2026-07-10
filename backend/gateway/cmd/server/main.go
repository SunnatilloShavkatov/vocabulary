package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

	vocabRepo := vocabularyrepository.NewVocabularyPgxRepository(pool)
	vocabularySvc := vocabularyservice.NewVocabularyService(cfg, vocabRepo)
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
	router := gwhttp.NewGatewayRouter(cfg.JWT.Secret, cfg.CORSAllowedOrigins, vocabularySvc, usersSvc, notificationSvc, authGRPCClient)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, router.Handler()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
