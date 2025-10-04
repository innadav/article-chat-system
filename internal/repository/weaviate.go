package repository

import (
	"context"
	"fmt"
	"log"
	"strings"

	"article-chat-system/internal/models"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	weaviate_models "github.com/weaviate/weaviate/entities/models"
)

const ArticleClassName = "Article"

type VectorRepository struct {
	client *weaviate.Client
}

func NewVectorRepository(host, scheme string) (*VectorRepository, error) {
	// The client config now correctly uses separate fields for Scheme and Host.
	cfg := weaviate.Config{
		Host:   host,
		Scheme: scheme,
	}
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create weaviate client: %w", err)
	}

	repo := &VectorRepository{client: client}
	if err := repo.ensureSchemaExists(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure weaviate schema: %w", err)
	}
	return repo, nil
}

func (r *VectorRepository) ensureSchemaExists(ctx context.Context) error {
	exists, err := r.client.Schema().ClassExistenceChecker().WithClassName(ArticleClassName).Do(ctx)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Weaviate class %s already exists", ArticleClassName)
		return nil
	}

	classObj := &weaviate_models.Class{
		Class:      ArticleClassName,
		Vectorizer: "text2vec-transformers", // Use the built-in vectorizer
		Properties: []*weaviate_models.Property{
			{Name: "url", DataType: []string{"text"}},
			{Name: "title", DataType: []string{"text"}},
			{Name: "summary", DataType: []string{"text"}},
			{Name: "excerpt", DataType: []string{"text"}},
			{Name: "sentiment", DataType: []string{"text"}},
			{Name: "topics", DataType: []string{"text[]"}},
			{Name: "entities", DataType: []string{"text[]"}},
		},
	}

	err = r.client.Schema().ClassCreator().WithClass(classObj).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate class: %w", err)
	}

	log.Printf("Created Weaviate class: %s", ArticleClassName)
	return nil
}

// SaveArticle lets Weaviate create the vector automatically from the content.
func (r *VectorRepository) SaveArticle(ctx context.Context, art *models.Article) error {
	// Use URL as the object ID for consistency
	objectID := strings.ReplaceAll(art.URL, "/", "_")
	objectID = strings.ReplaceAll(objectID, ":", "_")
	objectID = strings.ReplaceAll(objectID, ".", "_")

	properties := map[string]interface{}{
		"url":       art.URL,
		"title":     art.Title,
		"summary":   art.Summary,
		"excerpt":   art.Excerpt,
		"sentiment": art.Sentiment,
		"topics":    art.Topics,
		"entities":  art.Entities,
	}

	_, err := r.client.Data().Creator().
		WithClassName(ArticleClassName).
		WithID(objectID).
		WithProperties(properties).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to save article to Weaviate: %w", err)
	}

	log.Printf("Saved article to Weaviate: %s", art.Title)
	return nil
}

// SearchSimilarArticles finds relevant articles using a text query.
func (r *VectorRepository) SearchSimilarArticles(ctx context.Context, queryText string, limit int) ([]*models.Article, error) {
	nearText := r.client.GraphQL().NearTextArgBuilder().WithConcepts([]string{queryText})
	fields := []graphql.Field{
		{Name: "url"},
		{Name: "title"},
		{Name: "summary"},
		{Name: "excerpt"},
		{Name: "sentiment"},
		{Name: "topics"},
		{Name: "entities"},
	}

	result, err := r.client.GraphQL().Get().
		WithClassName(ArticleClassName).
		WithFields(fields...).
		WithNearText(nearText).
		WithLimit(limit).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to search articles in Weaviate: %w", err)
	}

	var articles []*models.Article
	data := result.Data["Get"].(map[string]interface{})
	items := data[ArticleClassName].([]interface{})

	for _, item := range items {
		itemMap := item.(map[string]interface{})

		article := &models.Article{
			URL:       getString(itemMap["url"]),
			Title:     getString(itemMap["title"]),
			Summary:   getString(itemMap["summary"]),
			Excerpt:   getString(itemMap["excerpt"]),
			Sentiment: getString(itemMap["sentiment"]),
		}

		// Parse topics array
		if topics, ok := itemMap["topics"].([]interface{}); ok {
			for _, topic := range topics {
				if topicStr, ok := topic.(string); ok {
					article.Topics = append(article.Topics, topicStr)
				}
			}
		}

		// Parse entities array
		if entities, ok := itemMap["entities"].([]interface{}); ok {
			for _, entity := range entities {
				if entityStr, ok := entity.(string); ok {
					article.Entities = append(article.Entities, entityStr)
				}
			}
		}

		articles = append(articles, article)
	}

	log.Printf("Found %d similar articles for query: %s", len(articles), queryText)
	return articles, nil
}

// getString safely extracts string value from interface{}
func getString(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
