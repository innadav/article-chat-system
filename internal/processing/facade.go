package processing

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/models"

	"github.com/go-shiori/go-readability"
)

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
	llmClient llm.Client
}

func NewAnalyzer(llmClient llm.Client) *Analyzer {
	return &Analyzer{llmClient: llmClient}
}

func (a *Analyzer) InitialAnalysis(ctx context.Context, art *models.Article) error {
	log.Printf("Performing initial analysis for %s", art.Title)

	// Generate summary
	summaryPrompt := fmt.Sprintf("Please provide a concise summary of this article titled '%s':\n\n%s", art.Title, art.Excerpt)
	summaryResp, err := a.llmClient.GenerateContent(ctx, summaryPrompt)
	if err != nil {
		log.Printf("Failed to generate summary for %s: %v", art.Title, err)
		art.Summary = fmt.Sprintf("Summary for %s: %s", art.Title, art.Excerpt)
	} else {
		art.Summary = summaryResp.Text
		log.Printf("Successfully generated summary for %s", art.Title)
	}

	// Extract topics/keywords
	topicsPrompt := fmt.Sprintf("Extract key topics and keywords from this article. Focus on the main themes, technologies, companies, and concepts mentioned:\n\nTitle: %s\nContent: %s\n\nReturn a concise list of 5-10 most relevant keywords and topics, separated by commas.", art.Title, art.Excerpt)
	topicsResp, err := a.llmClient.GenerateContent(ctx, topicsPrompt)
	if err != nil {
		log.Printf("Failed to extract topics for %s: %v", art.Title, err)
		art.Topics = []string{}
	} else {
		// Parse topics from response (comma-separated)
		topics := parseTopicsFromResponse(topicsResp.Text)
		art.Topics = topics
		log.Printf("Successfully extracted %d topics for %s", len(topics), art.Title)
	}

	// Analyze sentiment
	sentimentPrompt := fmt.Sprintf("Analyze the sentiment of this text and return only a number between -1 (very negative) and 1 (very positive):\n%s", art.Excerpt)
	sentimentResp, err := a.llmClient.GenerateContent(ctx, sentimentPrompt)
	if err != nil {
		log.Printf("Failed to analyze sentiment for %s: %v", art.Title, err)
		art.Sentiment = "neutral"
	} else {
		art.Sentiment = sentimentResp.Text
		log.Printf("Successfully analyzed sentiment for %s: %s", art.Title, art.Sentiment)
	}

	return nil
}

// parseTopicsFromResponse parses comma-separated topics from LLM response
func parseTopicsFromResponse(response string) []string {
	// Simple parsing - split by comma and clean up
	topics := []string{}
	for _, topic := range strings.Split(response, ",") {
		topic = strings.TrimSpace(topic)
		if topic != "" {
			topics = append(topics, topic)
		}
	}
	return topics
}

// Facade provides a simplified interface to the article processing subsystem.
type Facade struct {
	fetcher    *Fetcher
	analyzer   *Analyzer
	articleSvc article.Service
}

// NewFacade initializes the Facade with all its required subsystem components.
func NewFacade(llmClient llm.Client, articleSvc article.Service) *Facade {
	return &Facade{
		fetcher:    NewFetcher(),
		analyzer:   NewAnalyzer(llmClient),
		articleSvc: articleSvc,
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

	log.Printf("FACADE: Successfully processed and stored new article: %s", newArticle.Title)
	return newArticle, nil
}
