package main

import (
	"log"

	"github.com/Judgoo/JudgeX/server"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := server.Create()
	defer server.Release()
	if err := server.Listen(app); err != nil {
		log.Panic(err)
	}
}
