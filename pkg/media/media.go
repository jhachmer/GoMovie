package media

type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

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

// Movie struct for fields in an OMDb API response
type Movie struct {
	Media
}

type Series struct {
	Media
	TotalSeasons string `json:"totalSeasons"`
}

// Rating is a nested object in an OMDb Response
