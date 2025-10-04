package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"article-chat-system/internal/article"
	"article-chat-system/internal/cache"

	// "article-chat-system/internal/models"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/processing"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/repository"
	"article-chat-system/internal/strategies"
	"article-chat-system/internal/vector"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler now depends on the interfaces, not the concrete structs.
type Handler struct {
	logger           *slog.Logger         // Structured logger
	articleSvc       article.Service      // Use interface
	plannerSvc       planner.Service      // Use interface
	strategyExecutor *strategies.Executor // Executor can be concrete if it has no interface, or use its interface
	promptFactory    *prompts.Factory     // Factory can be concrete
	processingFacade *processing.Facade   // Facade can be concrete
	vectorSvc        vector.Service       // Vector service for semantic search
	cacheSvc         *cache.Service       // Cache service for API-level caching
}

// NewHandler now accepts the interfaces as arguments.
func NewHandler(
	logger *slog.Logger,
	articleSvc article.Service,
	plannerSvc planner.Service,
	strategyExecutor *strategies.Executor,
	promptFactory *prompts.Factory,
	processingFacade *processing.Facade,
	vectorSvc vector.Service,
	cacheSvc *cache.Service,
) *Handler {
	return &Handler{
		logger:           logger,
		articleSvc:       articleSvc,
		plannerSvc:       plannerSvc,
		strategyExecutor: strategyExecutor,
		promptFactory:    promptFactory,
		processingFacade: processingFacade,
		vectorSvc:        vectorSvc,
		cacheSvc:         cacheSvc,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Post("/chat", h.handleChat)
	r.Post("/articles", h.handleAddArticle)
	r.Post("/entities", h.handleFindEntities)
	return r
}

type ChatRequest struct {
	Query   string `json:"query"`
	Message string `json:"message"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
}

type AddArticleRequest struct {
	URL string `json:"url"`
}

type FindEntitiesRequest struct {
	URLs []string `json:"urls"`
}

type FindEntitiesResponse struct {
	Entities []repository.EntityCount `json:"entities"`
	Count    int                      `json:"count"`
}

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Support both "query" and "message" fields for backward compatibility
	query := req.Query
	if query == "" {
		query = req.Message
	}
	if query == "" {
		http.Error(w, "Either 'query' or 'message' field is required", http.StatusBadRequest)
		return
	}

	// --- NEW CACHING LOGIC ---
	// 1. Generate the cache key from the raw query string immediately.
	cacheKey := h.cacheSvc.GenerateCacheKey(query)

	// 2. Check the cache BEFORE any LLM calls.
	if cachedAnswer, found := h.cacheSvc.Get(cacheKey); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Answer: "🤖 (from cache)\n\n" + cachedAnswer})
		return // Return immediately on a cache hit.
	}
	// --- END NEW CACHING LOGIC ---

	// --- CACHE MISS: Proceed with the normal flow ---
	// 3. Create a plan (First LLM call).
	plan, err := h.plannerSvc.CreatePlan(r.Context(), query)
	if err != nil {
		h.logger.Error("Failed to create a query plan", "error", err, "raw_query", query)
		http.Error(w, "Failed to create a query plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Execute the plan (Potential second LLM call).
	answer, err := h.strategyExecutor.ExecutePlan(r.Context(), plan, h.articleSvc, h.promptFactory, h.vectorSvc)
	if err != nil {
		h.logger.Error("Failed to execute the plan", "error", err, "plan", plan)
		http.Error(w, "Failed to execute the plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Store the newly generated answer in the cache.
	h.cacheSvc.Set(cacheKey, answer)

	// 6. Return the response to the user.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Answer: answer})
}

func (h *Handler) handleAddArticle(w http.ResponseWriter, r *http.Request) {
	var req AddArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for add article", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	newArticle, err := h.processingFacade.AddNewArticle(r.Context(), req.URL)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.logger.Warn("Article already exists", "url", req.URL)
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.Error("Failed to process article", "error", err, "url", req.URL)
		http.Error(w, "Failed to process article: "+err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Info("Successfully processed new article", "url", req.URL, "title", newArticle.Title)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newArticle)
}

func (h *Handler) handleFindEntities(w http.ResponseWriter, r *http.Request) {
	var req FindEntitiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for find entities", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entities, err := h.articleSvc.FindCommonEntities(r.Context(), req.URLs)
	if err != nil {
		h.logger.Error("Failed to find common entities", "error", err, "urls", req.URLs)
		http.Error(w, "Failed to find common entities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully found common entities", "count", len(entities), "urls", req.URLs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FindEntitiesResponse{
		Entities: entities,
		Count:    len(entities),
	})
}
