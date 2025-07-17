package main

import (
	"log"
	"net/http"

	"github.com/AugustSerenity/marketplace/internal/handler"
	"github.com/AugustSerenity/marketplace/internal/service"
	"github.com/AugustSerenity/marketplace/internal/storage"
)

const portNumber = ":8080"

func main() {
	db := storage.InitDB()
	defer storage.CloseDB(db)

	st := storage.New(db)

	authService := service.New(st, "hello-Baddy")

	h := handler.New(authService, "hello-Baddy")

	s := http.Server{
		Addr:    portNumber,
		Handler: h.Route(),
	}

	log.Printf("Server started at http://localhost%s", portNumber)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
