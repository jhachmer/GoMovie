package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/jhachmer/gomovie/internal/config"
	"github.com/jhachmer/gomovie/internal/store"
	"github.com/jhachmer/gomovie/internal/util"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

const WatchedCol = 0
const TitleCol = 1
const YearCol = 2
const AddedByCol = 4

type CSVEntry struct {
	Watched bool
	Title   string
	Year    int
	AddedBy string
}

func (e CSVEntry) String() string {
	return fmt.Sprintf("Watched: %v, Title: %v, Year: %v", e.Watched, e.Title, e.Year)
}

func ReadCSV(reader io.Reader) ([]*CSVEntry, error) {
	records, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return nil, err
	}
	var entries []*CSVEntry
	for _, record := range records {
		entry, err := UnmarshalEntry(record)
		if err != nil {
			slog.Error("Error unmarshalling entry", "entry", entry, "err", err)
			continue
		}
		slog.Info("Found entry", "entry", entry)
		entries = append(entries, entry)
	}
	return entries, nil
}

func main() {
	reader, err := os.Open("movie_list.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer util.CloseOrLog(reader)
	_ = setup(reader)
}

func UnmarshalEntry(record []string) (*CSVEntry, error) {
	year, err := strconv.Atoi(record[YearCol])
	if err != nil {
		slog.Error("Error parsing year", "record", record, "year", year, "err", err)
		return nil, err
	}
	entry := &CSVEntry{
		Watched: watchedCheckboxValue(record[WatchedCol]),
		Title:   record[TitleCol],
		Year:    year,
		AddedBy: record[AddedByCol],
	}
	return entry, nil
}

func watchedCheckboxValue(value string) bool {
	return value == "TRUE"
}

type App struct {
	store       *store.Store
	CSVContents []*CSVEntry
}

func setup(reader io.Reader) *App {
	cfg := config.Envs
	fmt.Printf("%#v\n", cfg)
	dbStore, err := store.SetupDatabase(cfg)
	if err != nil {
		slog.Error("Setup database failed", "err", err.Error())
		os.Exit(1)
	}
	records, err := ReadCSV(reader)
	if err != nil {
		slog.Error("Reading CSV failed", "err", err.Error())
		os.Exit(1)
	}
	return &App{
		store:       &dbStore,
		CSVContents: records,
	}
}
