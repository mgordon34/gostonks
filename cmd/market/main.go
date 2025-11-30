package main

import (
	"log"
	"os"
)

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	log.Printf("starting market service; redis host=%s port=%s", redisHost, redisPort)
}
