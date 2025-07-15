package main

import (
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

	srv := service.New(st)

	h := handler.New(srv)

	s := http.Server{
		Addr:    portNumber,
		Handler: h.Route(),
	}

	s.ListenAndServe()
}
