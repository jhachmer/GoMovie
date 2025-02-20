package store

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jhachmer/gomovie/internal/types"
)

type EntryStore interface {
	CreateEntry(entry *types.Entry, movie *types.Movie) (*types.Entry, error)
	GetEntries(userID string) ([]*types.Entry, error)
	UpdateEntry(entryID, field, newValue string, watched bool) (*types.Entry, error)
	DeleteEntry(entryID string) error
}

func (s *SQLiteStorage) CreateEntry(e *types.Entry, mov *types.Movie) (*types.Entry, error) {
	var exists bool
	row := s.DB.QueryRow( /*sql*/ `
		SELECT EXISTS(SELECT media.title
		FROM media
		WHERE media.id = ?);
		`, mov.ImdbID)
	if err := row.Scan(&exists); err != nil {
		log.Println("movie exists:", exists)
		return nil, err
	} else if !exists {
		_, err := s.CreateMovie(mov)
		if err != nil {
			return nil, err
		}
	}
	var watchedInt = 0
	if e.Watched {
		watchedInt = 1
	}
	res, err := s.DB.Exec( /*sql*/ `
		INSERT INTO entries
		(name, watched, comment, media_id)
		VALUES (?, ?, ?, ?);
		`, e.Name, watchedInt, e.Comment, mov.ImdbID)
	if err != nil {
		return nil, err
	}
	e.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *SQLiteStorage) CreateEntryTx(tx *sql.Tx, e *types.Entry, mov *types.Movie) (*types.Entry, error) {
	var exists bool
	row := tx.QueryRow( /*sql*/ `
		SELECT EXISTS(SELECT media.title
		FROM media
		WHERE media.id = ?);
		`, mov.ImdbID)
	if err := row.Scan(&exists); err != nil {
		log.Println("movie exists:", exists)
		return nil, err
	} else if !exists {
		_, err := s.CreateMovieTx(tx, mov)
		if err != nil {
			return nil, err
		}
	}
	var watchedInt = 0
	if e.Watched {
		watchedInt = 1
	}
	res, err := tx.Exec( /*sql*/ `
		INSERT INTO entries
		(name, watched, comment, media_id)
		VALUES (?, ?, ?, ?);
		`, e.Name, watchedInt, e.Comment, mov.ImdbID)
	if err != nil {
		return nil, err
	}
	e.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *SQLiteStorage) UpdateEntry(movieId, name, comment string, watched bool) (*types.Entry, error) {
	var watchedInt = 0
	if watched {
		watchedInt = 1
	}
	res, err := s.DB.Exec( /*sql*/ `
		UPDATE entries
		SET name = ?, comment = ?, watched = ?
		WHERE media_id = ?
	`, name, comment, watchedInt, movieId)
	if err != nil {
		return nil, err
	}
	resID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	entry := types.Entry{
		ID:      resID,
		Name:    name,
		Comment: []byte(comment),
		Watched: watched,
	}
	return &entry, nil
}

func (s *SQLiteStorage) DeleteEntry(imdbId string) error {
	_, err := s.DB.Exec( /*sql*/ `
		DELETE FROM entries
		WHERE media_id = ?
	`, imdbId)
	if err != nil {
		return fmt.Errorf("error deleting entry for movie: %s\n%w", imdbId, err)
	}
	return nil
}

func (s *SQLiteStorage) GetEntries(id string) ([]*types.Entry, error) {
	rows, err := s.DB.Query( /*sql*/ `
		SELECT id, name, watched, comment
		FROM entries
		WHERE media_id = ?;
		`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*types.Entry

	for rows.Next() {
		var entry types.Entry
		if err := rows.Scan(&entry.ID, &entry.Name, &entry.Watched, &entry.Comment); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
