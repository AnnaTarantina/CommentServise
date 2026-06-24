package storage

func (s *DatabaseStorage) InitializeSchema() error {
	_, err := s.DB.Exec(`
        CREATE TABLE IF NOT EXISTS comments (
            id TEXT PRIMARY KEY,
            news_id TEXT NOT NULL,
            parent_id TEXT,
            text TEXT NOT NULL,
            author TEXT NOT NULL,
            created_at TEXT NOT NULL,
            is_approved BOOLEAN DEFAULT true
        )
    `)
	return err
}

func (s *DatabaseStorage) CheckConnection() error {
	return s.DB.Ping()
}
