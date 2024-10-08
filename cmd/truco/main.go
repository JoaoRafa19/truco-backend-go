package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/JoaoRafa19/truco-backend-go/internal/api"
	"github.com/JoaoRafa19/truco-backend-go/internal/store/pgstore"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("TRUCO_DATABASE_USER"),
		os.Getenv("TRUCO_DATABASE_PASSWORD"),
		os.Getenv("TRUCO_DATABASE_HOST"),
		os.Getenv("TRUCO_DATABASE_PORT"),
		os.Getenv("TRUCO_DATABASE_NAME"),
	))

	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	handler := api.NewHandler(pgstore.New(pool))

	go func() {
		fmt.Println(
			"PID: ",
			os.Getpid(),
		)
		fmt.Println("RUNNING ON: localhost:3000")
		if err := http.ListenAndServe(":3000", handler); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
}
