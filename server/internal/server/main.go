package server

import (
	"context"
	"log"

	"sideDesert/shiba/internal/server/controller"
	s "sideDesert/shiba/internal/server/services"
	vb "sideDesert/shiba/internal/vbrowser"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func init() {
	godotenv.Load(".env")
}

func NewServer(ctx context.Context, config *s.ServerConfig) (*controller.Controller, error) {
	// service, err := NewService(ctx, config)
	service, err := s.NewService(ctx, config)

	if err != nil {
		log.Print("Error in NewServer[NewService()]: ", err)
		return nil, err
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Print("Error in NewServer[NatsConnect()]: ", err)
		return nil, err
	}

	controller := controller.NewController(service, nc, vb.NewManager(99))

	return controller, nil
}
