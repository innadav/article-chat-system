package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"article-chat-system/internal/article"
	"article-chat-system/internal/cache"
	"article-chat-system/internal/config"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/processing"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/repository"
	"article-chat-system/internal/strategies"
	"article-chat-system/internal/tracing"
	handler "article-chat-system/internal/transport/http"
	"article-chat-system/internal/vector"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()

	// 1. Initialize the Tracer Provider at the very beginning.
	tp, err := tracing.InitTracerProvider()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// 2. Create a new structured JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Set it as the default logger for the whole application
	slog.SetDefault(logger)

	logger.Info("Starting article chat system", "version", "1.0.0")

	// 2. Load Configuration
	cfg := config.New()
	logger.Info("Configuration loaded",
		"llm_provider", cfg.LLMProvider,
		"prompt_version", cfg.PromptVersion,
		"database_url", cfg.DatabaseURL,
		"port", cfg.Port)

	// Initialize Database
	repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to initialize repository", "error", err)
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	// No need to defer db.Close() here, as it's handled by the repository
	logger.Info("Successfully connected to PostgreSQL")

	// 2. Initialize Core Components
	llmClient, err := llm.NewClientFactory(ctx, cfg)
	if err != nil {
		logger.Error("Failed to create LLM client", "error", err)
		log.Fatalf("Failed to create LLM client: %v", err)
	}
	logger.Info("LLM client created", "provider", cfg.LLMProvider)

	promptLoader, err := prompts.NewLoader(cfg.PromptVersion)
	if err != nil {
		logger.Error("Failed to load prompts", "error", err, "version", cfg.PromptVersion)
		log.Fatalf("Failed to load prompts: %v", err)
	}
	promptFactory, err := prompts.NewFactory(prompts.ModelGemini15Flash, promptLoader)
	if err != nil {
		logger.Error("Failed to create prompt factory", "error", err)
		log.Fatalf("Failed to create prompt factory: %v", err)
	}
	strategyExecutor := strategies.NewExecutor()

	// Initialize cache service for API-level caching
	cacheSvc := cache.NewService()
	logger.Info("Successfully initialized cache service")

	// Initialize vector repository (Weaviate)
	vecRepo, err := repository.NewVectorRepository(cfg.WeaviateHost, cfg.WeaviateScheme)
	if err != nil {
		logger.Warn("Failed to initialize Weaviate repository", "error", err, "host", cfg.WeaviateHost)
		logger.Info("Falling back to simple vector service")
		vecRepo = nil
	} else {
		logger.Info("Successfully initialized Weaviate repository", "host", cfg.WeaviateHost)
	}

	// 3. Initialize Services
	articleSvc := article.NewService(llmClient, repo, vecRepo)

	// Initialize vector service (fallback)
	var vectorSvc vector.Service
	if vecRepo == nil {
		vectorSvc = nil
		logger.Info("No vector service available")
	} else {
		// Use the existing Weaviate service as fallback
		weaviateSvc, err := vector.NewWeaviateService(cfg.WeaviateHost, cfg.WeaviateScheme, cfg.WeaviateAPIKey)
		if err != nil {
			logger.Warn("Failed to initialize Weaviate service", "error", err)
			vectorSvc = nil
		} else {
			logger.Info("Successfully initialized Weaviate service")
			vectorSvc = weaviateSvc
		}
	}

	plannerSvc := planner.NewService(llmClient, promptFactory, articleSvc, vecRepo)
	processingFacade := processing.NewFacade(llmClient, articleSvc, promptFactory, vectorSvc, vecRepo)

	// 4. Initialize the Transport Layer (The Handler) LAST
	apiHandler := handler.NewHandler(
		logger,
		articleSvc,
		plannerSvc,
		strategyExecutor,
		promptFactory,
		processingFacade,
		vectorSvc,
		cacheSvc,
	)

	// 5. Start Background Processes
	go func() {
		logger.Info("Processing initial articles in the background", "count", len(cfg.InitialArticleURLs))
		for _, url := range cfg.InitialArticleURLs {
			_, err := processingFacade.AddNewArticle(context.Background(), url)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					logger.Debug("Article already processed", "url", url)
				} else {
					logger.Error("Failed to process initial URL", "url", url, "error", err)
				}
			}
		}
		logger.Info("Initial article processing complete")
	}()

	// 6. Start the Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: otelhttp.NewHandler(apiHandler.Routes(), "http.server"),
	}
	go func() {
		logger.Info("Starting server", "port", "8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			log.Fatalf("Could not listen on :8080: %v\n", err)
		}
	}()

	// 7. Handle Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("Shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		log.Fatalf("Server shutdown failed: %v", err)
	}
	logger.Info("Server gracefully stopped")
}
