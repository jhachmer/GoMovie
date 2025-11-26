package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jhachmer/gomovie/internal/cache"
	"github.com/jhachmer/gomovie/internal/config"
	"github.com/jhachmer/gomovie/internal/handlers"
	"github.com/jhachmer/gomovie/internal/server"
	"github.com/jhachmer/gomovie/internal/store"
	"github.com/jhachmer/gomovie/internal/types"
)

func main() {
	checkForValidConfig()
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	_ = w
	_ = args
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg := config.Envs

	logger := log.New(os.Stdout, "gomovie:", log.LstdFlags)

	dbStore, err := store.SetupDatabase(cfg)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	svr := setupServer(dbStore, logger)
	err = svr.Serve(ctx)
	return err
}

func setupServer(store store.Store, logger *log.Logger) *server.Server {
	movC := cache.NewCache[string, *types.Movie](time.Second*15, time.Minute*60, nil)
	serC := cache.NewCache[string, *types.Series](time.Second*15, time.Minute*60, nil)
	handler := handlers.NewHandler(store, movC, serC, logger)

	return server.NewServer(config.Envs.Addr, logger, handler)
}

func checkForValidConfig() {
	if !config.Envs.Valid {
		log.Fatalln("Config is not valid! Check .env File for missing values")
	}
}
