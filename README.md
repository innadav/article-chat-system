# Article Chat System

A sophisticated chat-based service built in Go that enables users to interact with a collection of articles through natural language queries. The system provides intelligent analysis capabilities including summarization, topic extraction, sentiment analysis, and complex comparative analysis across multiple articles.

## 🏗️ Architecture Overview

The system is built on a clean, decoupled, and scalable architecture using several key design patterns:

- **Hexagonal Architecture**: Core application logic is isolated from external concerns like databases or web servers
- **Facade Pattern**: Complex article ingestion processes (fetching, parsing, analyzing, storing) are simplified behind a single `processing.Facade`
- **Strategy & Template Method Patterns**: Each chat query type is handled by a separate `Strategy` class with a `BaseStrategy` implementing common logic
- **Factory Pattern**: Centralized prompt engineering through `PromptFactory` and flexible LLM provider switching via `LLMFactory`

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Handler  │    │   Article Svc   │    │   Planner Svc   │
│                 │◄──►│                 │◄──►│                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Strategy Exec  │    │   PostgreSQL    │    │   LLM Client    │
│                 │    │   Repository    │    │   (OpenAI)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Processing     │    │   Weaviate      │    │   Cache Service │
│  Facade         │    │   Vector DB     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🔧 Key Design Decisions

### Observability & Monitoring
- **Structured Logging with `slog`**: All logs are written as structured JSON for machine readability and future monitoring integration
- **OpenTelemetry Integration**: Distributed tracing provides LangSmith-like visibility into every LLM call, showing prompts, responses, and timing
- **Jaeger Visualization**: Interactive trace analysis through Jaeger UI for debugging and performance optimization

### Data Storage & Retrieval
- **PostgreSQL**: Persistent storage for article metadata and pre-computed analyses
- **Weaviate Vector Database**: Efficient semantic search enabling scalable article discovery
- **In-Memory Caching**: API-level caching with request hashing for instant responses to repeated queries

### LLM Integration
- **Provider Abstraction**: Clean interface supporting multiple LLM providers (OpenAI, Mock)
- **Prompt Engineering**: Centralized prompt management with version control
- **Error Handling**: Comprehensive error tracking and graceful degradation

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- OpenAI API Key (or configure for other providers)

### Installation & Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/innadav/article-chat-system.git
   cd article-chat-system
   ```

2. **Set up environment variables**
   Create a `.env` file in the root directory:
   ```bash
   OPENAI_API_KEY="your-openai-api-key-here"
   GEMINI_API_KEY="your-gemini-api-key-here"  # Optional
   ```

3. **Start the services**
   ```bash
   docker-compose up --build
   ```

4. **Verify the setup**
   - **API**: http://localhost:8080
   - **Jaeger UI**: http://localhost:16686
   - **PostgreSQL**: localhost:5433
   - **Weaviate**: localhost:8081

## 📡 API Endpoints

### Chat Interface
```bash
POST /chat
Content-Type: application/json

{
  "query": "Summarize the main topics from the articles about AI"
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

## 🔍 Observability Features

### Structured Logging
All application logs are written in JSON format with contextual information:
```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "Successfully processed new article",
  "url": "https://example.com/article",
  "article_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

### Distributed Tracing
Every request creates a trace showing:
- HTTP request/response timing
- LLM call details (prompt, response, duration)
- Database operations
- Cache hits/misses
- Error propagation

View traces in Jaeger UI at http://localhost:16686

### Monitoring Integration Ready
The structured logging and tracing setup is designed for easy integration with:
- Prometheus metrics collection
- Grafana dashboards
- Alerting systems
- Log aggregation platforms (ELK stack, etc.)

## 🧪 Testing

### Run Unit Tests
```bash
go test ./...
```

### Run Integration Tests
```bash
./test_comprehensive.sh
```

### Test Coverage
```bash
go test -cover ./...
```

## 🏗️ Development

### Project Structure
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

### Adding New Strategies
1. Create a new strategy file in `internal/strategies/`
2. Implement the `Strategy` interface
3. Register the strategy in the executor
4. Add corresponding prompts in `configs/prompts/`

### Adding New LLM Providers
1. Implement the `Client` interface in `internal/llm/`
2. Add provider-specific tracing instrumentation
3. Update the factory in `internal/llm/factory.go`

## 🔧 Configuration

The system supports configuration through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | LLM provider (openai, mock) | `openai` |
| `OPENAI_API_KEY` | OpenAI API key | Required |
| `GEMINI_API_KEY` | Gemini API key | Optional |
| `DB_URL` | PostgreSQL connection string | Auto-configured |
| `WEAVIATE_HOST` | Weaviate server host | `weaviate:8080` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Jaeger endpoint | `http://jaeger:4317` |

## 📊 Performance Considerations

- **Caching**: API responses are cached to reduce LLM costs and latency
- **Vector Search**: Efficient semantic search scales to thousands of articles
- **Batch Processing**: Article ingestion processes multiple articles concurrently
- **Connection Pooling**: Database connections are pooled for optimal performance

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with Go's excellent standard library and ecosystem
- Uses OpenTelemetry for industry-standard observability
- Leverages Weaviate for vector search capabilities
- Integrates with OpenAI's powerful language models
