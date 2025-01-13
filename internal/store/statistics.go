package store

type WatchStats struct {
	NumOfWatched   int
	NumOfUnwatched int
	TotalMovies    int
}

func (s *SQLiteStorage) GetWatchCounts() (*WatchStats, error) {
	var stats WatchStats
	row := s.db.QueryRow( /*sql*/ `
	SELECT
    SUM(CASE WHEN watched = 1 THEN 1 ELSE 0 END) AS watched_count,
    SUM(CASE WHEN watched = 0 THEN 1 ELSE 0 END) AS unwatched_count,
    COUNT(*) AS total_movies
	FROM entries;
	`)
	err := row.Scan(&stats.NumOfWatched, &stats.NumOfUnwatched, &stats.TotalMovies)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
