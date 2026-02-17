package types

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/jhachmer/gomovie/internal/config"
	"github.com/jhachmer/gomovie/internal/util"
)

const IMDbIDPattern = `^tt\d{7,8}$`

type Validator interface {
	Validate() error
}

type MediaResponse interface {
	Validator
}
type APIResponse struct {
	Response string `json:"Response"`
}

func (a APIResponse) Validate() error {
	if a.Response != "True" {
		return fmt.Errorf("response value is false")
	}
	return nil
}

// MediaRequest is an interface that defines the methods required for interacting with the OMDB API.
// - SendRequest: Sends a request to the OMDB API and returns the corresponding movie details or an error.
// - Validate: Performs validation on the request parameters to ensure they meet the requirements of the OMDB API.
type MediaRequest interface {
	SendRequest() (any, error)
	Validator
}

// MovieIDRequest is a request using the movie IMDb ID
type MovieIDRequest struct {
	imdbID string
}

// NewMovieIDRequest returns a pointer to a new MovieIDRequest
func NewMovieIDRequest(imdbID string) (*MovieIDRequest, error) {
	req := MovieIDRequest{
		imdbID: imdbID,
	}
	return &req, nil
}

// SendRequest returns movie data of movie in MovieIDRequest
func (r MovieIDRequest) SendRequest() (any, error) {
	return GetMediaFromRequest[Movie](r)
}

// Validate validates IMDB id in request
// ID must have:
// - 7 or 8 digits
// - two leading tt characters
func (r MovieIDRequest) Validate() error {
	if !regexp.MustCompile(IMDbIDPattern).MatchString(r.imdbID) {
		return fmt.Errorf("id %s is not a valid id", r.imdbID)
	}
	return nil
}

// MovieTitleRequest is a request using the movie title and year
type MovieTitleRequest struct {
	title string
	year  string
}

// NewMovieTitleRequest returns a pointer to a new MovieTitleRequest
func NewMovieTitleRequest(title, year string) (*MovieTitleRequest, error) {
	req := MovieTitleRequest{
		title: title,
		year:  year,
	}
	return &req, nil
}

// SendRequest returns movie data of movie in MovieTitleRequest
func (r MovieTitleRequest) SendRequest() (any, error) {
	return GetMediaFromRequest[Movie](r)
}

// Validate validates title and year in request
// Title must not be empty
// Year must be 4 digits
func (r MovieTitleRequest) Validate() error {
	return validateTitle(r.title, r.year)
}

type SeriesIDRequest struct {
	imdbID string
}

func (s SeriesIDRequest) SendRequest() (any, error) {
	return GetMediaFromRequest[Series](s)

}

func (s SeriesIDRequest) Validate() error {
	return validateID(s.imdbID)
}

type SeriesTitleRequest struct {
	title string
	year  string
}

func (s SeriesTitleRequest) SendRequest() (any, error) {
	return GetMediaFromRequest[Series](s)
}

func (s SeriesTitleRequest) Validate() error {
	return validateTitle(s.title, s.year)
}

func validateTitle(title, year string) error {
	if len(title) < 1 {
		return fmt.Errorf("title %s is not valid", title)
	}
	if len(year) != 4 {
		return fmt.Errorf("year %s is not a valid year", year)
	}
	return nil
}

func validateID(id string) error {
	if !regexp.MustCompile(`^tt\d{7,8}$`).MatchString(id) {
		return fmt.Errorf("id %s is not a valid id", id)
	}
	return nil
}

func GetMediaFromRequest[M MediaType](r MediaRequest) (*M, error) {
	var m M
	requestURL, err := buildRequestURL(r)
	if err != nil {
		return nil, err
	}
	switch r.(type) {
	case MovieIDRequest, MovieTitleRequest:
		movie, err := UnmarshalResponse[Movie](requestURL)
		if err != nil {
			return nil, err
		}
		m = any(movie).(M)
	}
	return &m, nil
}

// buildRequestURL is building request URL depending on request type
// OMDB_KEY environment variable must be set
// id requests use i=id query
// title requests are using t=title and y=year queries
func buildRequestURL(r MediaRequest) (string, error) {
	if err := r.Validate(); err != nil {
		return "", fmt.Errorf("request not valid %w", err)
	}
	reqURL, err := url.Parse(fmt.Sprintf("http://www.omdbapi.com/?apikey=%s", config.Envs.OmdbApiKey))
	if err != nil {
		return "", err
	}
	switch v := r.(type) {
	case MovieTitleRequest:
		values := reqURL.Query()
		values.Add("type", "movie")
		values.Add("t", v.title)
		values.Add("y", v.year)
		reqURL.RawQuery = values.Encode()
		return reqURL.String(), nil
	case MovieIDRequest:
		values := reqURL.Query()
		values.Add("type", "movie")
		values.Add("i", v.imdbID)
		reqURL.RawQuery = values.Encode()
		return reqURL.String(), nil
	case SeriesTitleRequest:
		values := reqURL.Query()
		values.Add("type", "series")
		values.Add("t", v.title)
		values.Add("y", v.year)
		reqURL.RawQuery = values.Encode()
		return reqURL.String(), nil
	case SeriesIDRequest:
		values := reqURL.Query()
		values.Add("type", "series")
		values.Add("i", v.imdbID)
		reqURL.RawQuery = values.Encode()
		return reqURL.String(), nil
	default:
		return "", fmt.Errorf("no valid request type")
	}
}

func UnmarshalResponse[MT MediaType](requestURL string) (MT, error) {
	var media MT

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return media, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return media, err
	}
	defer res.Body.Close()
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return media, err
	}
	var apiResponse APIResponse
	err = util.UnmarshalTo(responseBody, &apiResponse)
	if err != nil {
		return media, err
	}
	if err := apiResponse.Validate(); err != nil {
		return media, err
	}
	err = util.UnmarshalTo(responseBody, &media)
	if err != nil {
		return media, err
	}
	return media, nil
}

type SearchQueryRequest struct {
	ImdbID string
	Title  string
	Year   string
	Type   string
}

type SearchResultMedia struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	ImdbID string `json:"imdbID"`
	Type   string `json:"Type"`
	Poster string `json:"Poster"`
}

type SearchResults struct {
	Search       []SearchResultMedia `json:"Search"`
	TotalResults string              `json:"totalResults"`
	Response     string              `json:"Response"`
}
