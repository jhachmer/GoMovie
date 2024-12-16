package media

import (
	"fmt"
	"github.com/jhachmer/gotocollection/pkg/config"
	"github.com/jhachmer/gotocollection/pkg/util"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type OmdbRequest interface {
	SendRequest() (*Movie, error)
	Validate() error
}

type OmdbIDRequest struct {
	imdbID string
}

type OmdbTitleRequest struct {
	title string
	year  string
}

func NewOmdbIDRequest(imdbID string) (*OmdbIDRequest, error) {
	req := OmdbIDRequest{
		imdbID: imdbID,
	}
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func NewOmdbTitleRequest(title, year string) (*OmdbTitleRequest, error) {
	req := OmdbTitleRequest{
		title: title,
		year:  year,
	}
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r OmdbTitleRequest) SendRequest() (*Movie, error) {
	requestURL, err := makeRequestURL(r)
	if err != nil {
		return nil, err
	}
	var mov Movie
	mov, err = decodeRequest(requestURL)
	if err != nil {
		return nil, err
	}
	return &mov, nil
}

func (r OmdbIDRequest) SendRequest() (*Movie, error) {
	requestURL, err := makeRequestURL(r)
	if err != nil {
		return nil, err
	}
	var mov Movie
	mov, err = decodeRequest(requestURL)
	if err != nil {
		return nil, err
	}
	return &mov, nil
}

func (r OmdbTitleRequest) Validate() error {
	if len(r.title) < 1 {
		return fmt.Errorf("title %s is not valid", r.title)
	}
	if len(r.year) != 4 {
		return fmt.Errorf("year %s is not a valid year", r.year)
	}
	return nil
}

func (r OmdbIDRequest) Validate() error {
	if !regexp.MustCompile(`tt\d{7,8}$`).MatchString(r.imdbID) {
		return fmt.Errorf("id %s is not a valid id", r.imdbID)
	}
	return nil
}

func makeRequestURL(r OmdbRequest) (string, error) {
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
		if err != nil {
			return "", err
		}
		reqURL.RawQuery = values.Encode()
		return reqURL.String(), nil
	case OmdbIDRequest:
		values := reqURL.Query()
		values.Add("i", v.imdbID)
		if err != nil {
			return "", err
		}
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
