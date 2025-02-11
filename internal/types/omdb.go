package types

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/jhachmer/gomovie/internal/config"
	"github.com/jhachmer/gomovie/internal/util"
)

// OmdbRequest is an interface that defines the methods required for interacting with the OMDB API.
// - SendRequest: Sends a request to the OMDB API and returns the corresponding movie details or an error.
// - Validate: Performs validation on the request parameters to ensure they meet the requirements of the OMDB API.
type OmdbRequest interface {
	SendRequest() (*Movie, error)
	Validate() error
}

// OmdbIDRequest is a request using the movies IMDb ID
type OmdbIDRequest struct {
	imdbID string
}

// OmdbTitleRequest is a request using the movies title and year
type OmdbTitleRequest struct {
	title string
	year  string
}

// NewOmdbIDRequest returns a pointer to a new OmdbIDRequest
func NewOmdbIDRequest(imdbID string) (*OmdbIDRequest, error) {
	req := OmdbIDRequest{
		imdbID: imdbID,
	}
	return &req, nil
}

// NewOmdbTitleRequest returns a pointer to a new OmdbTitleRequest
func NewOmdbTitleRequest(title, year string) (*OmdbTitleRequest, error) {
	req := OmdbTitleRequest{
		title: title,
		year:  year,
	}
	return &req, nil
}

// SendRequest returns movie data of movie in OmdbTitleRequest
func (r OmdbTitleRequest) SendRequest() (*Movie, error) {
	return sendAndReturn(r)
}

// SendRequest returns movie data of movie in OmdbIDRequest
func (r OmdbIDRequest) SendRequest() (*Movie, error) {
	return sendAndReturn(r)
}

func sendAndReturn(r OmdbRequest) (*Movie, error) {
	requestURL, err := makeRequestURL(r)
	if err != nil {
		return nil, err
	}
	var mov Movie
	mov, err = decodeRequest(requestURL)
	if err != nil {
		return nil, err
	}
	if mov.Response == "False" {
		return nil, fmt.Errorf("could not find movie with id %s", mov.Response)
	}
	return &mov, nil
}

// Validate validates title and year in request
// Title must not be empty
// Year must be 4 digits
func (r OmdbTitleRequest) Validate() error {
	if len(r.title) < 1 {
		return fmt.Errorf("title %s is not valid", r.title)
	}
	if len(r.year) != 4 {
		return fmt.Errorf("year %s is not a valid year", r.year)
	}
	return nil
}

// Validate validates IMDB id in request
// ID must have:
// - 7 or 8 digits
// - two leading tt characters
func (r OmdbIDRequest) Validate() error {
	if !regexp.MustCompile(`^tt\d{7,8}$`).MatchString(r.imdbID) {
		return fmt.Errorf("id %s is not a valid id", r.imdbID)
	}
	return nil
}

// makeRequestURL is building request URL depending on request type
// OMDB_KEY environment variable must be set
// id requests use i=id query
// title requests are using t=title and y=year queries
func makeRequestURL(r OmdbRequest) (string, error) {
	if err := r.Validate(); err != nil {
		return "", fmt.Errorf("request not valid %w", err)
	}
	if config.Envs.OmdbApiKey == "" {
		return "", fmt.Errorf("OMDb API is not set")
	}
	reqURL, err := url.Parse(fmt.Sprintf("http://www.omdbapi.com/?apikey=%s", config.Envs.OmdbApiKey))
	if err != nil {
		return "", err
	}
	switch v := r.(type) {
	case OmdbTitleRequest:
		values := reqURL.Query()
		values.Add("t", v.title)
		values.Add("y", v.year)
		reqURL.RawQuery = values.Encode()
		return reqURL.String(), nil
	case OmdbIDRequest:
		values := reqURL.Query()
		values.Add("i", v.imdbID)
		reqURL.RawQuery = values.Encode()
		return reqURL.String(), nil
	default:
		return "", fmt.Errorf("no valid request type")
	}
}

func decodeRequest(requestURL string) (Movie, error) {
	var mov Movie
	req, err := getRequest(requestURL)
	if err != nil {
		return mov, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return mov, err
	}
	mov, err = util.Decode[Movie](res)
	if err != nil {
		return mov, err
	}
	if !checkIfResponseTrue(mov) {
		return Movie{}, fmt.Errorf("response value is false")
	}
	return mov, nil
}

func getRequest(requestURL string) (*http.Request, error) {
	res, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func checkIfResponseTrue(mov Movie) bool {
	return strings.ToLower(mov.Response) == "true"
}
