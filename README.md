# Article Chat System

This project is a chat-based service, written in Go, that allows users to interact with a set of articles through a natural language interface. It can provide summaries, extract topics, analyze sentiment, and perform complex comparative analysis on a persistent collection of articles.

## Architecture Overview

The system is built on a clean, decoupled architecture using several key design patterns to separate concerns. There are two primary workflows: the **Article Ingestion Flow** and the **Chat Request Flow**.

### Article Ingestion Flow (Facade Pattern)

When a new article is added via the `/articles` endpoint, it follows this simplified process managed by the `processing.Facade`, which hides the complexity of fetching, analyzing, and storing the data.

```mermaid
graph TD
    A[HTTP POST /articles] --> B{processing.Facade};
    B --> C[1. Fetcher: Fetch & Parse URL];
    C --> D[2. Analyzer: Initial Analysis (LLM Call)];
    D --> E[3. Save to PostgreSQL];
    D --> F[4. Save to Weaviate];
```

### Chat Request Flow (Strategy Pattern)

When a user sends a query to the `/chat` endpoint, the system uses a multi-step process to generate an intelligent answer.

```mermaid
graph TD
    A[HTTP POST /chat] --> B{planner.Service};
    B --> C[1. Vector Search (Weaviate)];
    C --> D[2. Create Plan (LLM Call)];
    D --> E{strategies.Executor};
    E --> F[3. Select Strategy (e.g., Summarize)];
    F --> G[4. Execute (Final LLM Call)];
    G --> H[HTTP Response];
```

## Key Design Decisions

  - **Hexagonal Architecture**: The core application logic in the `internal` directory is isolated from external concerns. The `llm.Client` interface allows swapping AI providers, and the `repository.ArticleRepository` interface decouples the application from PostgreSQL.

  - **Facade Pattern**: Used in `internal/processing/facade.go` to provide a simple, single-method interface (`AddNewArticle`) for the complex, multi-step process of ingesting a new article.

  - **Strategy & Template Method Patterns**: Used in `internal/strategies/` to manage each chat query type as an interchangeable algorithm. This makes the system modular and easy to extend. A `BaseStrategy` provides a shared workflow skeleton.

  - **Factory Patterns**: The `llm.Factory` allows the system to select different LLM clients, and the `prompts.Factory` centralizes all prompt engineering by loading versioned templates from external YAML files.

  - **Repository Pattern**: Used in `internal/repository/` to abstract all database interactions for both PostgreSQL (metadata) and Weaviate (vector search).

  - **Observability**: The system is instrumented with **`slog`** for structured logging and **OpenTelemetry** for distributed tracing, providing LangSmith-like visibility into LLM calls via Jaeger.

## How to Run Locally (with Docker)

### Prerequisites

  - Docker and Docker Compose
  - A Gemini or OpenAI API Key

### 1\. Configure Environment

Create a `.env` file in the root of the project. This file will hold your secret API key and provider choice.

```
# .env file

# Set your provider ("openai" or "google")
LLM_PROVIDER="openai"

# Add the corresponding API key
OPENAI_API_KEY="sk-..."
GEMINI_API_KEY="AIza..."
```

### 2\. Run the Application

Use Docker Compose to build and run all the services (Go API, PostgreSQL, Weaviate, and Jaeger).

```bash
docker-compose up --build
```

### 3\. Access the Services

  - **API**: `http://localhost:8080`
  - **Jaeger UI (for Tracing)**: `http://localhost:16686`
  - **Weaviate UI**: `http://localhost:8081`

### 4\. Test the API

#### Add a New Article

```bash
curl -X POST http://localhost:8080/articles \
-H "Content-Type: application/json" \
-d '{
    "url": "https://techcrunch.com/2024/05/21/google-confirms-an-internal-documents-leak-detailing-how-its-search-ranking-works/"
}'
```

#### Ask for a Summary

```bash
curl -X POST http://localhost:8080/chat \
-H "Content-Type: application/json" \
-d '{
    "query": "summarize the article about the google documents leak"
}'
```