# Article Chat System

[![Go Report Card](https://goreportcard.com/badge/github.com/innadav/article-chat-system)](https://goreportcard.com/report/github.com/innadav/article-chat-system)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An intelligent, chat-based service built in Go that allows users to interact with a persistent collection of articles. The system leverages Large Language Models (LLMs) and a vector database to provide summarization, sentiment analysis, and complex, context-aware query responses.

## Architecture & System Flow

The system is built on a clean, decoupled architecture using several key design patterns to separate concerns.

### System Components

This diagram shows the main components and their dependencies. The `Handler` is the entry point, which uses the `Facade` for ingestion and the `Planner`/`Executor` for chat. All core logic uses the `ArticleService` to interact with the data layers.

```mermaid
graph TD
    subgraph "Transport Layer"
        A[HTTP Handler]
    end
    subgraph "Business Logic"
        B[Facade]
        C[Planner]
        D[Executor]
        E[Strategies]
        F[ArticleService]
    end
    subgraph "Data Layer"
        G[Postgres Repo]
        H[Weaviate Repo]
    end
    A --> B; A --> C; A --> D;
    B --> F; C --> F; D --> E; E --> F;
    F --> G; F --> H;
```

## Key Design Decisions

  - **Hexagonal Architecture**: Core application logic is isolated from external concerns. The `llm.Client` and `repository.ArticleRepository` interfaces allow swapping external services without changing business logic.

  - **Facade Pattern**: Used in `internal/processing/facade.go` to provide a simple, single-method interface (`AddNewArticle`) for the complex process of ingesting a new article.

  - **Strategy & Template Method Patterns**: Used in `internal/strategies/` to manage each chat query type as an interchangeable algorithm, making the system highly extensible. A `BaseStrategy` provides a shared workflow skeleton.

  - **Factory Patterns**:

      - The `llm.Factory` allows the system to select different LLM clients (e.g., Google Gemini, OpenAI). The system has been tested with **OpenAI's GPT-3.5 Turbo**, but can be easily switched to another model by changing the configuration.
      - The `prompts.Factory` centralizes all prompt engineering by loading versioned templates from external YAML files, separating prompt logic from business logic.

  - **API-Level Caching**: A simple **in-memory cache** (`go-cache`) with request hashing provides instant responses for repeated queries.

      - **Note**: For this proof-of-concept, the cache is in-memory and local to each container instance. For a production, multi-instance deployment, this would be replaced with a **distributed cache like Redis** to ensure consistency and persistence.

## 🚀 Getting Started

### Prerequisites

  - Docker and Docker Compose
  - An OpenAI or Gemini API Key

### 1\. Configure Environment

Create a `.env` file in the root of the project with the following structure:

```bash
# .env file

# Server Configuration
PORT=8080

# OpenAI Configuration
OPENAI_API_KEY=sk-proj-your_openai_api_key_here
OPENAI_MODEL=gpt-3.5-turbo

# Database Configuration
DATABASE_URL=postgres://user:password@postgres:5432/articledb?sslmode=disable
```

**Required Environment Variables:**
- `OPENAI_API_KEY`: Your OpenAI API key (required)
- `PORT`: Server port (default: 8080)
- `OPENAI_MODEL`: OpenAI model to use (default: gpt-3.5-turbo)
- `DATABASE_URL`: PostgreSQL connection string for the database

**Note:** The DATABASE_URL uses `postgres:5432` for container-to-container communication within Docker.

### 2\. Run the Application

Use Docker Compose to build and run all services.

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