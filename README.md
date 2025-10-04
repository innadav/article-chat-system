# Article Chat System

This project is a chat-based service, written in Go, that allows users to interact with a set of articles through a natural language interface. It can provide summaries, extract topics, analyze sentiment, and perform complex comparative analysis on a persistent collection of articles.

## Architecture Overview

The system is built on a clean, decoupled architecture using several key design patterns to separate concerns. There are two primary workflows: the **Article Ingestion Flow** and the **Chat Request Flow**.

### Article Ingestion Flow (Facade Pattern)

When a new article is added via the `/articles` endpoint, it follows this simplified process managed by the `processing.Facade`.

1.  **HTTP Request**: The `transport.http.Handler` receives a URL.
2.  **Facade**: The `processing.Facade` orchestrates the entire workflow.
3.  **Fetcher**: The `processing.Fetcher` retrieves and parses the article content.
4.  **Analyzer**: The `processing.Analyzer` performs an initial analysis (summary, keywords) by calling the configured LLM.
5.  **Persistence**: The final, enriched `models.Article` object is saved to the **PostgreSQL** database via the `repository.PostgresRepository`.

### Chat Request Flow (Strategy Pattern)

When a user sends a query to the `/chat` endpoint, the system uses a multi-step process to generate an intelligent answer.

1.  **Planning**: The `planner.Service` receives the user's query and uses an LLM to convert it into a structured `QueryPlan`, identifying the user's `intent`.
2.  **Routing**: The `strategies.Executor` receives the `QueryPlan` and selects the appropriate `Strategy` class (e.g., `SummarizeStrategy`, `CompareToneStrategy`) from its map.
3.  **Execution**: The chosen `Strategy` executes. It uses a `BaseStrategy` (Template Method Pattern) to handle common logic, while implementing its own specific logic. It uses the `prompts.Factory` to build a request for the LLM and the `article.Service` to retrieve data from the repository.
4.  **Response**: The final, synthesized answer from the LLM is returned to the user.

## Key Design Decisions

  - **Hexagonal Architecture (Ports & Adapters)**: The core application logic in the `internal` directory is completely isolated from external concerns. For example, the `llm.Client` interface allows swapping AI providers without changing business logic, and the `repository.ArticleRepository` interface decouples the application from PostgreSQL.

  - **Facade Pattern**: Used in `internal/processing/facade.go` to provide a simple, single-method interface (`AddNewArticle`) for the complex, multi-step process of ingesting a new article.

  - **Strategy & Template Method Patterns**: Used in `internal/strategies/` to manage each chat query type as an interchangeable algorithm. This makes the system extremely modular and easy to extend with new capabilities. The `BaseStrategy` provides a shared workflow skeleton.

  - **Factory Patterns**:

      * The `internal/llm/factory.go` allows the system to select and initialize different LLM clients (e.g., Gemini, OpenAI) based on configuration.
      * The `internal/prompts/factory.go` centralizes all prompt engineering, loading versioned templates from YAML files and separating prompt logic from business logic.

  - **Repository Pattern**: Used in `internal/repository/` to abstract all database interactions. This decouples the application from the specific database technology (PostgreSQL) and centralizes data access logic.

## Observability & Monitoring

The system includes enterprise-grade observability features:

- **Structured Logging**: All logs are written as structured JSON using Go's `slog` package for machine readability
- **Distributed Tracing**: OpenTelemetry integration provides LangSmith-like visibility into every LLM call
- **Jaeger Visualization**: Interactive trace analysis through Jaeger UI at http://localhost:16686
- **Performance Monitoring**: Detailed timing information for optimization and debugging
- **Token Usage Tracking**: Automatic token counting and cost monitoring (when API supports it)

## How to Run Locally (with Docker)

### Prerequisites

  - Docker and Docker Compose
  - A Gemini API Key (or OpenAI API Key)

### 1\. Setup

Clone the repository to your local machine:

```bash
git clone https://github.com/innadav/article-chat-system.git
cd article-chat-system
```

### 2\. Configure Environment

Create a `.env` file in the root of the project. This file will hold your secret API key.

```
# .env file
GEMINI_API_KEY="YOUR_API_KEY_HERE"
# Or if you are using OpenAI:
# OPENAI_API_KEY="YOUR_API_KEY_HERE"
```

### 3\. Run the Application

Use Docker Compose to build and run all the services (Go API, PostgreSQL, Weaviate, Jaeger, etc.).

```bash
docker-compose up --build
```

The API will be available at `http://localhost:8080`.

### 4\. Access Services

- **API**: http://localhost:8080
- **Jaeger UI**: http://localhost:16686 (for trace visualization)
- **PostgreSQL**: localhost:5433
- **Weaviate**: localhost:8081

### 5\. Test the API

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

#### Find Common Entities

```bash
curl -X POST http://localhost:8080/entities \
-H "Content-Type: application/json" \
-d '{
    "urls": ["https://example.com/article1", "https://example.com/article2"]
}'
```

## API Endpoints

### Chat Interface
```bash
POST /chat
Content-Type: application/json

{
  "query": "What are the main topics in the articles about AI?"
}
```

### Add New Article
```bash
POST /articles
Content-Type: application/json

{
  "url": "https://example.com/article"
}
```

### Find Common Entities
```bash
POST /entities
Content-Type: application/json

{
  "urls": ["https://example.com/article1", "https://example.com/article2"]
}
```

## Project Structure

```
├── cmd/server/           # Application entry point
├── internal/
│   ├── article/         # Article domain logic
│   ├── cache/           # Caching service
│   ├── config/          # Configuration management
│   ├── llm/             # LLM client abstractions
│   ├── planner/         # Query planning service
│   ├── processing/      # Article processing facade
│   ├── prompts/         # Prompt management
│   ├── repository/      # Data access layer
│   ├── strategies/      # Query execution strategies
│   ├── tracing/         # OpenTelemetry setup
│   └── transport/http/  # HTTP handlers
├── tests/               # Test suites
└── configs/prompts/     # Prompt templates
```

## Configuration

The system supports configuration through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | LLM provider (openai, gemini) | `openai` |
| `OPENAI_API_KEY` | OpenAI API key | Required |
| `GEMINI_API_KEY` | Gemini API key | Optional |
| `DB_URL` | PostgreSQL connection string | Auto-configured |
| `WEAVIATE_HOST` | Weaviate server host | `weaviate:8080` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Jaeger endpoint | `jaeger:4317` |

## Testing

### Run Unit Tests
```bash
go test ./...
```

### Run Integration Tests
```bash
./test_comprehensive.sh
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.