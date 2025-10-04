package repository

import (
	"article-chat-system/internal/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// PostgresRepository is the concrete implementation for PostgreSQL.
type PostgresRepository struct {
	DB *sql.DB
}

// NewPostgresRepository creates a new repository and pings the database.
func NewPostgresRepository(dbURL string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}
	return &PostgresRepository{DB: db}, nil
}

// NewPostgresRepositoryWithDB creates a new repository with an existing database connection.
func NewPostgresRepositoryWithDB(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{DB: db}
}

// Save inserts or updates an article in the database.
func (r *PostgresRepository) Save(ctx context.Context, art *models.Article) error {
	query := `
		INSERT INTO articles (url, title, excerpt, summary, sentiment, topics, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (url) DO UPDATE SET
			title = EXCLUDED.title,
			excerpt = EXCLUDED.excerpt,
			summary = EXCLUDED.summary,
			sentiment = EXCLUDED.sentiment,
			topics = EXCLUDED.topics;
	`
	_, err := r.DB.ExecContext(ctx, query,
		art.URL, art.Title, art.Excerpt,
		art.Summary, art.Sentiment, pq.Array(art.Topics), art.ProcessedAt,
	)
	return err
}

// FindByURL retrieves an article by its URL.
func (r *PostgresRepository) FindByURL(ctx context.Context, url string) (*models.Article, error) {
	var art models.Article
	query := `SELECT url, title, excerpt, summary, sentiment, topics, processed_at FROM articles WHERE url = $1`
	err := r.DB.QueryRowContext(ctx, query, url).Scan(
		&art.URL, &art.Title, &art.Excerpt,
		&art.Summary, &art.Sentiment, pq.Array(&art.Topics), &art.ProcessedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("error finding article by URL: %w", err)
	}
	return &art, nil
}

// FindAll retrieves all articles.
func (r *PostgresRepository) FindAll(ctx context.Context) ([]*models.Article, error) {
	query := `SELECT url, title, excerpt, summary, sentiment, topics, processed_at FROM articles`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error finding all articles: %w", err)
	}
	defer rows.Close()

	var articles []*models.Article
	for rows.Next() {
		var art models.Article
		if err := rows.Scan(
			&art.URL, &art.Title, &art.Excerpt,
			&art.Summary, &art.Sentiment, pq.Array(&art.Topics), &art.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning article: %w", err)
		}
		articles = append(articles, &art)
	}
	return articles, rows.Err()
}
