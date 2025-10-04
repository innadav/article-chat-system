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
		INSERT INTO articles (url, title, excerpt, summary, sentiment, topics, entities, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (url) DO UPDATE SET
			title = EXCLUDED.title,
			excerpt = EXCLUDED.excerpt,
			summary = EXCLUDED.summary,
			sentiment = EXCLUDED.sentiment,
			topics = EXCLUDED.topics,
			entities = EXCLUDED.entities;
	`
	_, err := r.DB.ExecContext(ctx, query,
		art.URL, art.Title, art.Excerpt,
		art.Summary, art.Sentiment, pq.Array(art.Topics), pq.Array(art.Entities), art.ProcessedAt,
	)
	return err
}

// FindByURL retrieves an article by its URL.
func (r *PostgresRepository) FindByURL(ctx context.Context, url string) (*models.Article, error) {
	var art models.Article
	query := `SELECT url, title, excerpt, summary, sentiment, topics, entities, processed_at FROM articles WHERE url = $1`
	err := r.DB.QueryRowContext(ctx, query, url).Scan(
		&art.URL, &art.Title, &art.Excerpt,
		&art.Summary, &art.Sentiment, pq.Array(&art.Topics), pq.Array(&art.Entities), &art.ProcessedAt,
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
	query := `SELECT url, title, excerpt, summary, sentiment, topics, entities, processed_at FROM articles`
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
			&art.Summary, &art.Sentiment, pq.Array(&art.Topics), pq.Array(&art.Entities), &art.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning article: %w", err)
		}
		articles = append(articles, &art)
	}
	return articles, rows.Err()
}

// FindTopEntities finds the most common entities across articles using efficient PostgreSQL query
func (r *PostgresRepository) FindTopEntities(ctx context.Context, articleURLs []string, limit int) ([]EntityCount, error) {
	var query string
	var args []interface{}

	if len(articleURLs) == 0 {
		// Query all articles - prioritize entities over topics
		query = `
			WITH entity_counts AS (
				SELECT unnest(entities) as entity, COUNT(*) as count
				FROM articles 
				WHERE entities IS NOT NULL AND array_length(entities, 1) > 0
				GROUP BY entity
				UNION ALL
				SELECT unnest(topics) as entity, COUNT(*) as count
				FROM articles 
				WHERE topics IS NOT NULL AND array_length(topics, 1) > 0
				AND (entities IS NULL OR array_length(entities, 1) = 0)
				GROUP BY entity
			)
			SELECT entity, SUM(count) as total_count
			FROM entity_counts
			GROUP BY entity
			ORDER BY total_count DESC, entity ASC
			LIMIT $1
		`
		args = []interface{}{limit}
	} else {
		// Query specific articles by URL
		query = `
			WITH entity_counts AS (
				SELECT unnest(entities) as entity, COUNT(*) as count
				FROM articles 
				WHERE url = ANY($1) AND entities IS NOT NULL AND array_length(entities, 1) > 0
				GROUP BY entity
				UNION ALL
				SELECT unnest(topics) as entity, COUNT(*) as count
				FROM articles 
				WHERE url = ANY($1) AND topics IS NOT NULL AND array_length(topics, 1) > 0
				AND (entities IS NULL OR array_length(entities, 1) = 0)
				GROUP BY entity
			)
			SELECT entity, SUM(count) as total_count
			FROM entity_counts
			GROUP BY entity
			ORDER BY total_count DESC, entity ASC
			LIMIT $2
		`
		args = []interface{}{pq.Array(articleURLs), limit}
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error finding top entities: %w", err)
	}
	defer rows.Close()

	var entities []EntityCount
	for rows.Next() {
		var entity EntityCount
		if err := rows.Scan(&entity.Entity, &entity.Count); err != nil {
			return nil, fmt.Errorf("error scanning entity: %w", err)
		}
		entities = append(entities, entity)
	}

	return entities, rows.Err()
}
