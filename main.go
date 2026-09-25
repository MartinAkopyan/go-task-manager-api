package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	ctx := context.Background()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	r := chi.NewRouter()


	srv := &http.Server{Addr: ":" + os.Getenv("PORT"), Handler: r}

	pool, err := DBPool(ctx)

	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	log.Println("Succesfull connection")

	tHandler := TaskHandler{db: pool}

	r.Post("/tasks", tHandler.Create)
	r.Get("/tasks", tHandler.ListTasks)
	r.Get("/tasks/{id}", tHandler.GetTask)
	r.Patch("/tasks/{id}", tHandler.UpdateTask)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()


	<-signalChan
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("server stopped")

	defer pool.Close()

}
