package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	gwhttp "vocabulary/backend/gateway/internal/http"
	"vocabulary/backend/libs/shared/config"
	"vocabulary/backend/libs/shared/db"
	"vocabulary/backend/modules/auth/repository"
	"vocabulary/backend/modules/auth/service"
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
	authRepo := authrepository.NewAuthPgxRepository(pool)
	authSvc := authservice.NewAuthService(cfg, authRepo)
	vocabularySvc := vocabularyservice.NewVocabularyService(cfg, vocabRepo)
	router := gwhttp.NewGatewayRouter(cfg.JWT.Secret, authSvc, vocabularySvc)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, router.Handler()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
