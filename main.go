package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	pool, err := DBPool(ctx)

	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	defer pool.Close()

	log.Println("Succesfull connection")

}
