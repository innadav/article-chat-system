package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"article-chat-system/internal/article"
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
	articleSvc       article.Service      // Use interface
	plannerSvc       planner.Service      // Use interface
	strategyExecutor *strategies.Executor // Executor can be concrete if it has no interface, or use its interface
	promptFactory    *prompts.Factory     // Factory can be concrete
	processingFacade *processing.Facade   // Facade can be concrete
	vectorSvc        vector.Service       // Vector service for semantic search
}

// NewHandler now accepts the interfaces as arguments.
func NewHandler(
	articleSvc article.Service,
	plannerSvc planner.Service,
	strategyExecutor *strategies.Executor,
	promptFactory *prompts.Factory,
	processingFacade *processing.Facade,
	vectorSvc vector.Service,
) *Handler {
	return &Handler{
		articleSvc:       articleSvc,
		plannerSvc:       plannerSvc,
		strategyExecutor: strategyExecutor,
		promptFactory:    promptFactory,
		processingFacade: processingFacade,
		vectorSvc:        vectorSvc,
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

	plan, err := h.plannerSvc.CreatePlan(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to create a query plan: "+err.Error(), http.StatusInternalServerError)
		return
	}
	answer, err := h.strategyExecutor.ExecutePlan(r.Context(), plan, h.articleSvc, h.promptFactory, h.vectorSvc)
	if err != nil {
		http.Error(w, "Failed to execute the plan: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Answer: answer})
}

func (h *Handler) handleAddArticle(w http.ResponseWriter, r *http.Request) {
	var req AddArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	newArticle, err := h.processingFacade.AddNewArticle(r.Context(), req.URL)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Failed to process article: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newArticle)
}

func (h *Handler) handleFindEntities(w http.ResponseWriter, r *http.Request) {
	var req FindEntitiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entities, err := h.articleSvc.FindCommonEntities(r.Context(), req.URLs)
	if err != nil {
		http.Error(w, "Failed to find common entities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FindEntitiesResponse{
		Entities: entities,
		Count:    len(entities),
	})
}
