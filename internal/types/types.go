package types

import (
	"slices"
)

// Entry holds data regarding user submitted info
// e.g. username, if the movie is already watched and a comment
type Entry struct {
	ID      int64
	Name    string
	Watched bool
	Comment []byte
}

// NewEntry returns pointer to a new Entry instance
func NewEntry(name string, watched bool, comment string) *Entry {
	return &Entry{
		ID:      0,
		Name:    name,
		Watched: watched,
		Comment: []byte(comment),
	}
}

type MovieInfoData struct {
	Movie *Movie
	Entry []*Entry
}

type SeriesInfoData struct {
	Series *Series
	Entry  []*Entry
}

// SortMovieSlice sorts slice of movies based on their title
func SortMovieSlice(movies []*MovieInfoData) {
	slices.SortFunc(movies, func(a, b *MovieInfoData) int {
		if a.Movie.Title < b.Movie.Title {
			return -1
		}
		if a.Movie.Title > b.Movie.Title {
			return 1
		}
		return 0
	})
}

// MovieInfoPage holds necessary data for the InfoHandler
type MovieInfoPage struct {
	Entries []*Entry
	Movie   *Movie
	Error   error
}

type MovieOverviewData struct {
	Movies []*MovieInfoData
	Error  error
}

type SeriesOverviewData struct {
	Series []*SeriesInfoData
	Error  error
}

type LoginData struct {
	Error error
}

type SearchParams struct {
	Genres []string
	Actors []string
	Years  YearSearch
}

type YearSearch struct {
	StartYear string
	EndYear   string
}

type WatchStats struct {
	NumOfWatched   int
	NumOfUnwatched int
	TotalMovies    int
}
