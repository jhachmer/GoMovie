package main

import (
	"github.com/jhachmer/gotocollection/internal/cache"
	"github.com/jhachmer/gotocollection/internal/config"
	"github.com/jhachmer/gotocollection/internal/handlers"
	"github.com/jhachmer/gotocollection/internal/server"
	"github.com/jhachmer/gotocollection/internal/store"
	"github.com/jhachmer/gotocollection/internal/types"
	"log"
	"os"
	"time"
)

func main() {
	logger := log.New(os.Stdout, "goto:", log.LstdFlags)

	db, err := store.NewSQLiteStorage(config.Envs)
	if err != nil {
		log.Fatal(err)
	}
	dbStore := store.NewStore(db)
	dbStore.TestDBConnection()
	err = dbStore.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}
	movC := cache.NewCache[string, *types.Movie](time.Second*15, time.Second*60, nil)
	handler := handlers.NewHandler(dbStore, movC)

	svr := server.NewServer(config.Envs.Addr, logger, handler)
	svr.Serve()
}
