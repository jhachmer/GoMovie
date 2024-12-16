package main

import (
	"database/sql"
	"github.com/jhachmer/gotocollection/pkg/config"
	"github.com/jhachmer/gotocollection/pkg/server"
	"github.com/jhachmer/gotocollection/pkg/store"
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "goto:", log.LstdFlags)

	db, err := store.NewSQLiteStorage(config.Envs)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dbStore := store.NewStore(db)
	testDBConnection(db)
	err = dbStore.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}

	handler := server.NewHandler(dbStore)

	svr := server.NewServer(config.Envs.Addr, logger, handler)
	svr.Serve()
}

func testDBConnection(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to DB...")
}
