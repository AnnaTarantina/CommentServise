package storage

import (
	"database/sql"

	"../models"
)

type DatabaseStorage struct {
	DB *sql.DB
}

func NewDatabaseStorage(connectionString string) (*DatabaseStorage, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
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
	if err != nil {
		return nil, err
	}

	return &DatabaseStorage{DB: db}, nil
}

func (s *DatabaseStorage) SaveComment(comment *models.Comment) error {
	query := `
        INSERT INTO comments (id, news_id, parent_id, text, author, created_at, is_approved)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	_, err := s.DB.Exec(
		query,
		comment.ID,
		comment.NewsID,
		comment.ParentID,
		comment.Text,
		comment.Author,
		comment.CreatedAt,
		comment.IsApproved,
	)
	return err
}

func (s *DatabaseStorage) GetCommentsByNewsID(newsID string) ([]models.Comment, error) {
	var comments []models.Comment
	rows, err := s.DB.Query(`
        SELECT id, news_id, parent_id, text, author, created_at, is_approved
        FROM comments
        WHERE news_id = $1
        ORDER BY created_at
    `, newsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.NewsID,
			&comment.ParentID,
			&comment.Text,
			&comment.Author,
			&comment.CreatedAt,
			&comment.IsApproved,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}
