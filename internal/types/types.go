package types

import (
	"fmt"
	"slices"
)

// Rating holds rating data, which are pairs of Source and the actual rating value
type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

func (r Rating) String() string {
	return fmt.Sprint(r.Value)
}

// Media struct holds data acquired from omdb api
type Media struct {
	Title      string   `json:"Title"`
	Year       string   `json:"Year"`
	Rated      string   `json:"Rated"`
	Released   string   `json:"Released"`
	Runtime    string   `json:"Runtime"`
	Genre      string   `json:"Genre"`
	Director   string   `json:"Director"`
	Writer     string   `json:"Writer"`
	Actors     string   `json:"Actors"`
	Plot       string   `json:"Plot"`
	Language   string   `json:"Language"`
	Country    string   `json:"Country"`
	Awards     string   `json:"Awards"`
	Poster     string   `json:"Poster"`
	Ratings    []Rating `json:"Ratings,omitempty"`
	Metascore  string   `json:"Metascore,omitempty"`
	ImdbRating string   `json:"imdbRating,omitempty"`
	ImdbVotes  string   `json:"imdbVotes,omitempty"`
	ImdbID     string   `json:"imdbID"`
	Type       string   `json:"Type"`
	BoxOffice  string   `json:"BoxOffice"`
	Website    string   `json:"Website"`
	Response   string   `json:"Response"`
}

type Movie struct {
	Media
}

type Series struct {
	Media
	TotalSeasons string `json:"totalSeasons"`
}

// NewMovieFromID returns pointer to a new movie instance
// creates a new OmdbIDRequest and sends it to receive data
func NewMovieFromID(imdbID string) (*Movie, error) {
	req, err := NewOmdbIDRequest(imdbID)
	if err != nil {
		return nil, err
	}
	mov, err := req.SendRequest()
	if err != nil {
		return nil, err
	}
	return mov, nil
}

// NewMovieFromTitleAndYear returns pointer to a new movie instance
// creates a new OmdbTitleRequest and sends it to receive data
func NewMovieFromTitleAndYear(title, year string) (*Movie, error) {
	req, err := NewOmdbTitleRequest(title, year)
	if err != nil {
		return nil, err
	}
	mov, err := req.SendRequest()
	if err != nil {
		return nil, err
	}
	return mov, nil
}

func SortMovieSlice(movies []*Movie) {
	slices.SortFunc(movies, func(a, b *Movie) int {
		if a.Title < b.Title {
			return -1
		}
		if a.Title > b.Title {
			return 1
		}
		return 0
	})
}

// InfoPage holds necessary data for the InfoHandler
type InfoPage struct {
	Entries []*Entry
	Movie   *Movie
	Error   error
}

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

type HomeData struct {
	Movies []*Movie
	Error  error
}
