package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jhachmer/gomovie/internal/store"
	"github.com/jhachmer/gomovie/internal/types"
)

type FileParser struct {
	l      Logging
	file   *os.File
	reader *csv.Reader
	store  *store.SQLiteStorage
	*ParseArgs
}

type ParseArgs struct {
	Path             string
	TitleIndex       int
	YearIndex        int
	RecommenderIndex int
	WatchedIndex     int
}

type MovieAndEntry struct {
	mov   *types.Movie
	entry *types.Entry
}

type Logging struct {
	log *log.Logger
	f   *os.File
}

func setupLogging() (*Logging, error) {
	f, err := os.OpenFile("failed.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	logger := log.New(f, "gomovie Parser:", log.LstdFlags)
	return &Logging{
		log: logger,
		f:   f,
	}, nil
}

func DBForParsedContents() *store.SQLiteStorage {
	db, err := store.SetupDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	return db
}

// ParseCSV parses a CSV file and stores its contents into the database.
func ParseCSV(args *ParseArgs) error {
	logger, err := setupLogging()
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	defer logger.f.Close()

	if !strings.HasSuffix(args.Path, ".csv") {
		return fmt.Errorf("file parser currently only supports csv files")
	}
	if args.TitleIndex < 0 || args.YearIndex < 0 {
		return fmt.Errorf("indexes must be greater than zero")
	}
	f, err := os.Open(args.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	fp := FileParser{
		l:         *logger,
		file:      f,
		store:     DBForParsedContents(),
		ParseArgs: args,
	}
	fp.reader = csv.NewReader(fp.file)
	fp.reader.FieldsPerRecord = -1
	mae := fp.readMoviesAndEntries()
	err = fp.addContentsToDB(mae)
	if err != nil {
		return fmt.Errorf("error adding to db: %w", err)
	}
	return nil
}

func (fp *FileParser) readCSVContents() ([][]string, error) {
	fp.reader = csv.NewReader(fp.file)
	fp.reader.FieldsPerRecord = -1
	contents, err := fp.reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (fp *FileParser) addContentsToDB(mae []*MovieAndEntry) error {
	tx, err := fp.store.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Perform all inserts
	for _, e := range mae {
		_, err := fp.store.CreateMovieTx(tx, e.mov)
		if err != nil {
			return err
		}
		_, err = fp.store.CreateEntryTx(tx, e.entry, e.mov)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (fp *FileParser) readMoviesAndEntries() []*MovieAndEntry {
	parsedMovies := make([]*MovieAndEntry, 0)
	for {
		row, err := fp.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil
		}
		next, err := fp.processRow(row)
		if err != nil {
			fmt.Printf("%v\n", err)
			fp.l.log.Printf("%v\n", err)
			continue
		}
		parsedMovies = append(parsedMovies, next)
	}
	return parsedMovies
}

func (fp *FileParser) processRow(row []string) (*MovieAndEntry, error) {
	title := row[fp.TitleIndex]
	year := row[fp.YearIndex]
	mov, err := types.MovieFromTitleAndYear(title, year)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch movie: %s (%s)\n%w", title, year, err)
	}

	watched, err := strconv.ParseBool(row[fp.WatchedIndex])
	if err != nil {
		watched = false
	}

	entry := types.Entry{
		Name:    row[fp.RecommenderIndex],
		Watched: watched,
	}

	return &MovieAndEntry{mov: mov, entry: &entry}, nil
}
