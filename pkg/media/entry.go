package media

type Entry struct {
	ID      int64
	Name    string
	Watched bool
	Comment []byte
	Movie   *Movie
}

func NewEntryFromId(name, imdb, comment string, watched bool) (*Entry, error) {
	movReq, err := NewOmdbIDRequest(imdb)
	if err != nil {
		return nil, err
	}
	mov, err := movReq.SendRequest()
	if err != nil {
		return nil, err
	}
	return &Entry{
		Name:    name,
		Watched: watched,
		Comment: []byte(comment),
		Movie:   mov,
	}, nil
}

func NewEntryFromTitleAndYear(name, title, year, comment string, watched bool) (*Entry, error) {
	movReq, err := NewOmdbTitleRequest(title, year)
	if err != nil {
		return nil, err
	}
	mov, err := movReq.SendRequest()
	if err != nil {
		return nil, err
	}
	return &Entry{
		Name:    name,
		Watched: watched,
		Comment: []byte(comment),
		Movie:   mov,
	}, nil
}
