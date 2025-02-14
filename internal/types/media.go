package types

import (
	"fmt"
	"strings"
)

// Rating holds rating data, which are pairs of Source and the actual rating value
type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

// String method to implement Stringer interface
func (r Rating) String() string {
	return fmt.Sprint(r.Value)
}

// MediaType struct holds data acquired from omdb api
type MediaType interface {
	Movie | Series
	checkResponse() bool
}

type Movie struct {
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

func (m Movie) checkResponse() bool {
	return strings.ToLower(m.Response) == "true"
}

type Series struct {
	Movie
	TotalSeasons string `json:"totalSeasons"`
}

func (s Series) checkResponse() bool {
	return strings.ToLower(s.Response) == "true"
}

// MovieFromID returns pointer to a new movie instance
// creates a new MovieIDRequest and sends it to receive data
func MovieFromID(imdbID string) (*Movie, error) {
	req, err := NewMovieIDRequest(imdbID)
	if err != nil {
		return nil, err
	}
	res, err := req.SendRequest()
	if err != nil {
		return nil, err
	}
	mov, ok := res.(*Movie)
	if !ok {
		return nil, fmt.Errorf("type assertion for id %s failed", imdbID)
	}
	return mov, nil
}

// MovieFromTitleAndYear returns pointer to a new movie instance
// creates a new MovieTitleRequest and sends it to receive data
func MovieFromTitleAndYear(title, year string) (*Movie, error) {
	req, err := NewMovieTitleRequest(title, year)
	if err != nil {
		return nil, err
	}
	res, err := req.SendRequest()
	if err != nil {
		return nil, err
	}
	mov, ok := res.(*Movie)
	if !ok {
		return nil, fmt.Errorf("type assertion for id %s (%s) failed", title, year)
	}
	return mov, nil
}
