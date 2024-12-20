package media

// InfoPage holds necessary data for the InfoHandler
type InfoPage struct {
	Entries []*Entry
	Movie   *Movie
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

// NewInfoPage returns pointer to a new InfoPage instance
func NewInfoPage(mov *Movie, entries []*Entry) *InfoPage {
	return &InfoPage{
		Entries: entries,
		Movie:   mov,
	}
}
