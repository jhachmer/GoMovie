package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jhachmer/gotocollection/internal/cache"
	"github.com/jhachmer/gotocollection/internal/config"
	"github.com/jhachmer/gotocollection/internal/handlers"
	"github.com/jhachmer/gotocollection/internal/server"
	"github.com/jhachmer/gotocollection/internal/store"
	"github.com/jhachmer/gotocollection/internal/types"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := log.New(os.Stdout, "goto:", log.LstdFlags)

	dbStore, err := setupDatabase()

	svr := setupServer(dbStore, logger)
	err = svr.Serve(ctx)
	return err
}

func setupDatabase() (*store.Storage, error) {
	db, err := store.NewSQLiteStorage(config.Envs)
	if err != nil {
		return nil, err
	}
	dbStore := store.NewStore(db)
	dbStore.TestDBConnection()
	err = dbStore.InitDatabase()
	if err != nil {
		return nil, err
	}
	return dbStore, nil
}

func setupServer(store store.Store, logger *log.Logger) *server.Server {
	movC := cache.NewCache[string, *types.Movie](time.Second*15, time.Second*60, nil)
	handler := handlers.NewHandler(store, movC, logger)

	return server.NewServer(config.Envs.Addr, logger, handler)
}
