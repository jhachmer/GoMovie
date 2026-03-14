package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/jhachmer/go-cache"
	"github.com/jhachmer/gomovie/internal/api"
	"github.com/jhachmer/gomovie/internal/config"
	"github.com/jhachmer/gomovie/internal/handlers"
	"github.com/jhachmer/gomovie/internal/server"
	"github.com/jhachmer/gomovie/internal/store"
	"github.com/jhachmer/gomovie/internal/util"
)

func main() {
	debug := os.Getenv("DEBUG") != ""
	logLevel := new(slog.LevelVar)
	logLevel.Set(util.ResolveLogLevel(debug))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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

	cfg := config.Envs

	slog.Info("Welcome to gomovie", "goVersion", runtime.Version(), "goOS", runtime.GOOS, "goArch", runtime.GOARCH)

	dbStore, err := store.SetupDatabase(cfg)
	if err != nil {
		slog.Error("failed setting up db", "err", err.Error())
		cancel()
		os.Exit(1)
	}

	svr := setupServer(dbStore)
	err = svr.Serve(ctx)
	cancel()
	return err
}

func setupServer(store store.Store) *server.Server {
	movC := cache.NewTTLCache[string, *api.Movie](time.Second*15, time.Minute*60, nil)
	serC := cache.NewTTLCache[string, *api.Series](time.Second*15, time.Minute*60, nil)
	handler := handlers.NewHandler(store, movC, serC)

	return server.NewServer(config.Envs.Addr, handler)
}

func checkForValidConfig() {
	if !config.Envs.Valid {
		log.Fatalln("Config is not valid! Check .env File for missing values")
	}
}
