package processing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/models"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/repository"
	"article-chat-system/internal/vector"

	"github.com/go-shiori/go-readability"
)

// Define a struct to match the JSON output from the LLM for initial analysis.
type initialAnalysisResult struct {
	Headline  string   `json:"headline"`
	KeyPoints []string `json:"key_points"`
	Sentiment string   `json:"sentiment"`
	Entities  []string `json:"entities"`
}

type ParsedArticle struct {
	Title       string
	TextContent string
	Excerpt     string
}

type Fetcher struct{}

func NewFetcher() *Fetcher {
	return &Fetcher{}
}

func (f *Fetcher) FetchAndParse(ctx context.Context, url string) (*ParsedArticle, error) {
	// Placeholder: In a real app, use a proper HTTP client with context, timeouts, etc.
	articleData, err := readability.FromURL(url, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch or parse article: %w", err)
	}

	return &ParsedArticle{
		Title:       articleData.Title,
		TextContent: articleData.TextContent,
		Excerpt:     articleData.Excerpt,
	}, nil
}

type Analyzer struct {
	llmClient     llm.Client
	promptFactory *prompts.Factory
}

func NewAnalyzer(llmClient llm.Client, promptFactory *prompts.Factory) *Analyzer {
	return &Analyzer{
		llmClient:     llmClient,
		promptFactory: promptFactory,
	}
}

func (a *Analyzer) InitialAnalysis(ctx context.Context, art *models.Article) error {
	log.Printf("Performing comprehensive analysis for %s", art.Title)

	// Log input content size before creating prompt
	contentLength := len(art.TextContent)
	contentWords := len(strings.Fields(art.TextContent))
	estimatedTokens := contentLength / 4 // Rough estimate: 1 token ≈ 4 characters
	log.Printf("Input content size - Characters: %d, Words: %d, Estimated tokens: %d",
		contentLength, contentWords, estimatedTokens)

	// Use the new simpler initial analysis prompt for more reliable results
	prompt, err := a.promptFactory.CreateInitialAnalysisPrompt(art.TextContent)
	if err != nil {
		log.Printf("Failed to create initial analysis prompt for %s: %v", art.Title, err)
		return fmt.Errorf("failed to create initial analysis prompt: %w", err)
	}

	// Log final prompt size
	promptLength := len(prompt)
	promptWords := len(strings.Fields(prompt))
	promptTokens := promptLength / 4
	log.Printf("Final prompt size - Characters: %d, Words: %d, Estimated tokens: %d",
		promptLength, promptWords, promptTokens)

	resp, err := a.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		log.Printf("Failed to generate initial analysis for %s: %v", art.Title, err)
		return fmt.Errorf("failed to generate initial analysis: %w", err)
	}

	// Parse the JSON response
	var analysis initialAnalysisResult
	if err := json.Unmarshal([]byte(resp.Text), &analysis); err != nil {
		log.Printf("Failed to parse initial analysis JSON for %s: %v", art.Title, err)
		return fmt.Errorf("failed to parse initial analysis JSON: %w", err)
	}

	// Populate the main Article object with the richer data
	art.Summary = analysis.Headline + "\n- " + strings.Join(analysis.KeyPoints, "\n- ")
	art.Sentiment = analysis.Sentiment
	art.Topics = analysis.Entities
	art.Entities = analysis.Entities // Also populate entities field

	log.Printf("Successfully analyzed %s: %s", art.Title, analysis.Headline)
	return nil
}

// Facade provides a simplified interface to the article processing subsystem.
type Facade struct {
	fetcher    *Fetcher
	analyzer   *Analyzer
	articleSvc article.Service
	vectorSvc  vector.Service
	vecRepo    *repository.VectorRepository
}

// NewFacade initializes the Facade with all its required subsystem components.
func NewFacade(llmClient llm.Client, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service, vecRepo *repository.VectorRepository) *Facade {
	return &Facade{
		fetcher:    NewFetcher(),
		analyzer:   NewAnalyzer(llmClient, promptFactory),
		articleSvc: articleSvc,
		vectorSvc:  vectorSvc,
		vecRepo:    vecRepo,
	}
}

// AddNewArticle is the single method that hides the complex processing steps.
func (f *Facade) AddNewArticle(ctx context.Context, url string) (*models.Article, error) {
	log.Printf("FACADE: Starting to process new article from URL: %s", url)
	if _, ok := f.articleSvc.GetArticle(ctx, url); ok {
		return nil, fmt.Errorf("article from URL %s already exists", url)
	}

	// 1. Coordinate the Fetcher
	parsedArticle, err := f.fetcher.FetchAndParse(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetcher failed: %w", err)
	}

	newArticle := &models.Article{
		URL:         url,
		Title:       parsedArticle.Title,
		Excerpt:     parsedArticle.Excerpt,
		TextContent: parsedArticle.TextContent,
		ProcessedAt: time.Now(),
	}

	// 2. Coordinate the Analyzer
	if err := f.analyzer.InitialAnalysis(ctx, newArticle); err != nil {
		log.Printf("WARNING: Initial analysis failed for %s: %v", url, err)
	}

	// 3. Coordinate the Article Service to store the final result
	if err := f.articleSvc.StoreArticle(ctx, newArticle); err != nil {
		return nil, fmt.Errorf("failed to store article: %w", err)
	}

	// 4. Save content to Weaviate for vectorization and search (if available)
	if f.vecRepo != nil {
		if err := f.vecRepo.SaveArticle(ctx, newArticle); err != nil {
			log.Printf("WARNING: Failed to save article vector for %s: %v", url, err)
			// We can choose to not fail the whole operation if vectorization fails.
		}
	}

	// 5. Index the article in the vector database (if available)
	if f.vectorSvc != nil {
		if err := f.vectorSvc.IndexArticle(ctx, newArticle); err != nil {
			log.Printf("WARNING: Failed to index article in vector database: %v", err)
			// Don't fail the entire operation if vector indexing fails
		}
	}

	log.Printf("FACADE: Successfully processed and stored new article: %s", newArticle.Title)
	return newArticle, nil
}
