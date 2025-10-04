package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"article-chat-system/internal/article"
	"article-chat-system/internal/config"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/processing"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/repository"
	"article-chat-system/internal/strategies"
	handler "article-chat-system/internal/transport/http"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	// 1. Load Configuration
	cfg := config.New()
	log.Printf("Configuration loaded. LLM Provider: %s, Prompt Version: %s", cfg.LLMProvider, cfg.PromptVersion)

	// Initialize Database
	repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	// No need to defer db.Close() here, as it's handled by the repository
	log.Println("Successfully connected to PostgreSQL.")

	// 2. Initialize Core Components
	llmClient, err := llm.NewClientFactory(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}
	promptLoader, err := prompts.NewLoader(cfg.PromptVersion)
	if err != nil {
		log.Fatalf("Failed to load prompts: %v", err)
	}
	promptFactory, err := prompts.NewFactory(prompts.ModelGemini15Flash, promptLoader)
	if err != nil {
		log.Fatalf("Failed to create prompt factory: %v", err)
	}
	strategyExecutor := strategies.NewExecutor()

	// 3. Initialize Services
	articleSvc := article.NewService(llmClient, repo)
	plannerSvc := planner.NewService(llmClient, promptFactory, articleSvc)
	processingFacade := processing.NewFacade(llmClient, articleSvc)

	// 4. Initialize the Transport Layer (The Handler) LAST
	apiHandler := handler.NewHandler(
		articleSvc,
		plannerSvc,
		strategyExecutor,
		promptFactory,
		processingFacade,
	)

	// 5. Start Background Processes
	go func() {
		log.Println("Processing initial articles in the background...")
		for _, url := range cfg.InitialArticleURLs {
			_, err := processingFacade.AddNewArticle(context.Background(), url)
			if err != nil {
				log.Printf("Failed to process initial URL %s: %v", url, err)
			}
		}
		log.Println("Initial article processing complete.")
	}()

	// 6. Start the Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: apiHandler.Routes(),
	}
	go func() {
		log.Println("Starting server on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8080: %v\n", err)
		}
	}()

	// 7. Handle Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped.")
}
