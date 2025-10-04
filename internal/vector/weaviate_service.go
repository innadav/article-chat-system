package vector

import (
	"article-chat-system/internal/models"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

// WeaviateService implements vector operations using Weaviate
type WeaviateService struct {
	client  *weaviate.Client
	class   string
}

// NewWeaviateService creates a new Weaviate service
func NewWeaviateService(weaviateURL, apiKey string) (*WeaviateService, error) {
	config := weaviate.Config{
		Host:   weaviateURL,
		Scheme: "http",
	}

	if apiKey != "" {
		config.Headers = map[string]string{
			"X-API-Key": apiKey,
		}
	}

	client, err := weaviate.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Weaviate client: %w", err)
	}

	service := &WeaviateService{
		client: client,
		class:  "Article",
	}

	// Ensure the class exists
	if err := service.ensureClassExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure class exists: %w", err)
	}

	return service, nil
}

// ensureClassExists creates the Article class if it doesn't exist
func (w *WeaviateService) ensureClassExists() error {
	// Check if class exists
	exists, err := w.client.Schema().ClassExistenceChecker().WithClassName(w.class).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to check class existence: %w", err)
	}

	if exists {
		return nil
	}

	// Create the class
	class := &weaviate.Class{
		Class: w.class,
		Properties: []*weaviate.Property{
			{
				Name:     "url",
				DataType: []string{"string"},
			},
			{
				Name:     "title",
				DataType: []string{"string"},
			},
			{
				Name:     "excerpt",
				DataType: []string{"string"},
			},
			{
				Name:     "summary",
				DataType: []string{"string"},
			},
			{
				Name:     "sentiment",
				DataType: []string{"string"},
			},
			{
				Name:     "topics",
				DataType: []string{"string[]"},
			},
			{
				Name:     "entities",
				DataType: []string{"string[]"},
			},
			{
				Name:     "processedAt",
				DataType: []string{"date"},
			},
		},
		Vectorizer: "text2vec-openai", // Use OpenAI embeddings
		ModuleConfig: map[string]interface{}{
			"text2vec-openai": map[string]interface{}{
				"model": "ada-002",
			},
		},
	}

	err = w.client.Schema().ClassCreator().WithClass(class).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}

	log.Printf("Created Weaviate class: %s", w.class)
	return nil
}

// SearchByTopics searches for articles that contain the specified topics/parameters
func (w *WeaviateService) SearchByTopics(ctx context.Context, topics []string, limit int) ([]*models.Article, error) {
	if len(topics) == 0 {
		return []*models.Article{}, nil
	}

	// Create OR conditions for each topic
	var conditions []*filters.WhereBuilder
	for _, topic := range topics {
		// Search in title, summary, and topics
		titleCondition := filters.Where().
			WithPath([]string{"title"}).
			WithOperator(filters.Like).
			WithValueText("*" + topic + "*")
		
		summaryCondition := filters.Where().
			WithPath([]string{"summary"}).
			WithOperator(filters.Like).
			WithValueText("*" + topic + "*")
		
		topicCondition := filters.Where().
			WithPath([]string{"topics"}).
			WithOperator(filters.ContainsAny).
			WithValueString(topic)

		// OR condition for this topic
		topicOrCondition := filters.Where().
			WithOperator(filters.Or).
			WithOperands([]*filters.WhereBuilder{titleCondition, summaryCondition, topicCondition})
		
		conditions = append(conditions, topicOrCondition)
	}

	// Combine all topic conditions with OR
	var whereFilter *filters.WhereBuilder
	if len(conditions) == 1 {
		whereFilter = conditions[0]
	} else {
		whereFilter = filters.Where().
			WithOperator(filters.Or).
			WithOperands(conditions)
	}

	// Build the query
	builder := w.client.GraphQL().Get().
		WithClassName(w.class).
		WithFields(
			graphql.Field{Name: "url"},
			graphql.Field{Name: "title"},
			graphql.Field{Name: "excerpt"},
			graphql.Field{Name: "summary"},
			graphql.Field{Name: "sentiment"},
			graphql.Field{Name: "topics"},
			graphql.Field{Name: "entities"},
			graphql.Field{Name: "processedAt"},
		).
		WithWhere(whereFilter).
		WithLimit(limit)

	result, err := builder.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %w", err)
	}

	return w.parseSearchResults(result)
}

// SearchBySemanticSimilarity searches for articles semantically similar to the query
func (w *WeaviateService) SearchBySemanticSimilarity(ctx context.Context, query string, limit int) ([]*models.Article, error) {
	// Use nearText search for semantic similarity
	builder := w.client.GraphQL().Get().
		WithClassName(w.class).
		WithFields(
			graphql.Field{Name: "url"},
			graphql.Field{Name: "title"},
			graphql.Field{Name: "excerpt"},
			graphql.Field{Name: "summary"},
			graphql.Field{Name: "sentiment"},
			graphql.Field{Name: "topics"},
			graphql.Field{Name: "entities"},
			graphql.Field{Name: "processedAt"},
		).
		WithNearText(&graphql.NearTextArgument{
			Values: []string{query},
		}).
		WithLimit(limit)

	result, err := builder.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles by semantic similarity: %w", err)
	}

	return w.parseSearchResults(result)
}

// IndexArticle adds an article to the vector database
func (w *WeaviateService) IndexArticle(ctx context.Context, article *models.Article) error {
	// Convert article to Weaviate object
	weaviateObject := map[string]interface{}{
		"url":         article.URL,
		"title":       article.Title,
		"excerpt":     article.Excerpt,
		"summary":     article.Summary,
		"sentiment":   article.Sentiment,
		"topics":      article.Topics,
		"entities":    article.Entities,
		"processedAt": article.ProcessedAt,
	}

	// Use URL as the object ID for consistency
	objectID := strings.ReplaceAll(article.URL, "/", "_")
	objectID = strings.ReplaceAll(objectID, ":", "_")
	objectID = strings.ReplaceAll(objectID, ".", "_")

	_, err := w.client.Data().Creator().
		WithClassName(w.class).
		WithID(objectID).
		WithProperties(weaviateObject).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to index article: %w", err)
	}

	log.Printf("Indexed article in Weaviate: %s", article.Title)
	return nil
}

// RemoveArticle removes an article from the vector database
func (w *WeaviateService) RemoveArticle(ctx context.Context, url string) error {
	// Convert URL to object ID
	objectID := strings.ReplaceAll(url, "/", "_")
	objectID = strings.ReplaceAll(objectID, ":", "_")
	objectID = strings.ReplaceAll(objectID, ".", "_")

	err := w.client.Data().Deleter().
		WithClassName(w.class).
		WithID(objectID).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove article: %w", err)
	}

	log.Printf("Removed article from Weaviate: %s", url)
	return nil
}

// parseSearchResults converts Weaviate search results to Article models
func (w *WeaviateService) parseSearchResults(result *graphql.Response) ([]*models.Article, error) {
	var articles []*models.Article

	if result.Errors != nil {
		return nil, fmt.Errorf("graphql errors: %v", result.Errors)
	}

	data := result.Data["Get"].(map[string]interface{})
	articleList := data[w.class].([]interface{})

	for _, item := range articleList {
		articleData := item.(map[string]interface{})
		
		article := &models.Article{
			URL:     getString(articleData["url"]),
			Title:   getString(articleData["title"]),
			Excerpt: getString(articleData["excerpt"]),
			Summary: getString(articleData["summary"]),
			Sentiment: getString(articleData["sentiment"]),
		}

		// Parse topics array
		if topics, ok := articleData["topics"].([]interface{}); ok {
			for _, topic := range topics {
				if topicStr, ok := topic.(string); ok {
					article.Topics = append(article.Topics, topicStr)
				}
			}
		}

		// Parse entities array
		if entities, ok := articleData["entities"].([]interface{}); ok {
			for _, entity := range entities {
				if entityStr, ok := entity.(string); ok {
					article.Entities = append(article.Entities, entityStr)
				}
			}
		}

		articles = append(articles, article)
	}

	return articles, nil
}

// getString safely extracts string value from interface{}
func getString(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
