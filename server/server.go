package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/ferjoaguilar/rest-ws/database"
	"github.com/ferjoaguilar/rest-ws/repository"
	"github.com/ferjoaguilar/rest-ws/websocket"
	"github.com/gorilla/mux"
)

type Config struct {
	Port        string
	JWTSecret   string
	DatabaseUrl string
}

type Server interface {
	Config() *Config
	// WEboscket hub connection
	Hub() *websocket.Hub
}

type Broker struct {
	config *Config
	router mux.Router

	//hub websocket
	hub *websocket.Hub
}

func (b *Broker) Config() *Config {
	return b.config
}

func (b *Broker) Hub() *websocket.Hub {
	return b.hub
}

func NewServer(ctx context.Context, config *Config) (*Broker, error) {
	if config.Port == "" {
		return nil, errors.New("PORT is required")
	}

	if config.JWTSecret == "" {
		return nil, errors.New("JWTSecret is required")
	}

	if config.DatabaseUrl == "" {
		return nil, errors.New("DatabaseUrl is required")
	}

	broker := &Broker{
		config: config,
		router: *mux.NewRouter(),
		hub:    websocket.NewHub(),
	}

	return broker, nil
}

func (b *Broker) Start(binder func(s Server, r *mux.Router)) {
	b.router = *mux.NewRouter()
	binder(b, &b.router)

	repo, err := database.NewPostgresRepository(b.config.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	//websocket to server
	go b.hub.Run()

	repository.SetRepository(repo)

	log.Println("Starting server on port", b.Config().Port)
	if err := http.ListenAndServe(b.Config().Port, &b.router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
