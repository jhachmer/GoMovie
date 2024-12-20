package media

// Rating holds rating data, which are pairs of Source and the actual rating value
type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
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
