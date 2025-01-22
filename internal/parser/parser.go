package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jhachmer/gotocollection/internal/store"
	"github.com/jhachmer/gotocollection/internal/types"
)

type FileParser struct {
	file   *os.File
	reader *csv.Reader
	db     *store.SQLiteStorage
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

func DBForParsedContents() *store.SQLiteStorage {
	db, err := store.SetupDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	return db
}

func ParseCSV(args *ParseArgs) error {
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
		file:      f,
		db:        DBForParsedContents(),
		ParseArgs: args,
	}
	contents, err := fp.readCSVContents()
	if err != nil {
		return err
	}
	mae := fp.readMoviesAndEntries(contents)
	err = fp.addContentsToDB(mae)
	if err != nil {

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
	for _, e := range mae {
		_, err := fp.db.CreateMovie(e.mov)
		if err != nil {
			return fmt.Errorf("error adding to db movie: %s (%s)\n%w", e.mov.Title, e.mov.Year, err)
		}
		_, err = fp.db.CreateEntry(e.entry, e.mov)
		if err != nil {
			return fmt.Errorf("error creating entry for movie: %s (%s)\n%w", e.mov.Title, e.mov.Year, err)
		}
	}
	return nil
}

func (fp *FileParser) readMoviesAndEntries(contents [][]string) []*MovieAndEntry {
	parsedMovies := make([]*MovieAndEntry, 0)
	for _, row := range contents {
		title := row[fp.TitleIndex]
		year := row[fp.YearIndex]
		mov, err := types.NewMovieFromTitleAndYear(title, year)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to fetch movie %s (%s)\n", title, year)
			continue
		}
		watched, err := strconv.ParseBool(row[fp.WatchedIndex])
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to read bool %s\n", row[fp.WatchedIndex])
			watched = false
		}
		entry := types.Entry{
			Name:    row[fp.RecommenderIndex],
			Watched: watched,
		}
		mae := MovieAndEntry{
			mov:   mov,
			entry: &entry,
		}
		parsedMovies = append(parsedMovies, &mae)
	}
	return parsedMovies
}
