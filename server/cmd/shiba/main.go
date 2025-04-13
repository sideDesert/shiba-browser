package main

import (
	"context"
	"log"
	"os"
	"sideDesert/shiba/internal/server"
	"sideDesert/shiba/internal/server/services"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	var dbUrl = os.Getenv("DB_URL")

	config := &services.ServerConfig{
		DbUrl: dbUrl,
	}

	server, err := server.NewServer(ctx, config)

	if err != nil {
		panic(err)
	}

	defer server.CloseDbConn(ctx)
	server.Run(":9000")
	log.Println("✅ main() exited successfully")
}
