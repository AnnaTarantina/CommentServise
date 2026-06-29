package storage

import (
	"database/sql"

	"github.com/AnnaTarantina/CommentServise/models"
)

// DatabaseStorage обёртка над sql.DB для работы с комментариями
type DatabaseStorage struct {
	DB *sql.DB
}

// NewDatabaseStorage создаёт подключение к БД
func NewDatabaseStorage(connectionString string) (*DatabaseStorage, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DatabaseStorage{DB: db}, nil
}

// SaveComment сохраняет комментарий в БД
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

// GetCommentsByNewsID возвращает все комментарии для указанной новости
func (s *DatabaseStorage) GetCommentsByNewsID(newsID string) ([]models.Comment, error) {
	// Инициализируем слайс, чтобы в JSON возвращался [] вместо null
	comments := make([]models.Comment, 0)

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
		if err := rows.Scan(
			&comment.ID,
			&comment.NewsID,
			&comment.ParentID,
			&comment.Text,
			&comment.Author,
			&comment.CreatedAt,
			&comment.IsApproved,
		); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	// Проверяем ошибки итератора
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
