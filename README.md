# Flow

A polyglot AI chatbot platform with high-performance native modules, parallel LLM inference, and graph-based entity storage.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Flow Architecture                            │
├─────────────────────────────────────────────────────────────────────┤
│   Client (Web/Discord)                                               │
│        │                                                             │
│        ▼                                                             │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │                    Go Server (:8090)                         │   │
│   │  ┌───────────┬───────────┬───────────┬───────────────────┐  │   │
│   │  │  /chat    │ /pipeline │  /fast    │     /graph        │  │   │
│   │  │ WebSocket │  LLM Call │ Native Go │  DB Operations    │  │   │
│   │  └───────────┴───────────┴───────────┴───────────────────┘  │   │
│   └─────────────────────────────────────────────────────────────┘   │
│        │                 │                          │                │
│        │                 ▼                          ▼                │
│        │   ┌─────────────────────────┐   ┌──────────────────────┐   │
│        │   │  Python LLM (:8001)     │   │   PostgreSQL (:5432) │   │
│        │   │  ┌─────┬─────┬───────┐  │   │   ┌──────────────┐   │   │
│        │   │  │Date │Person│ Org  │  │   │   │ graph_nodes  │   │   │
│        │   │  │Worker│Worker│Worker│  │   │   │ graph_edges  │   │   │
│        │   │  └─────┴─────┴───────┘  │   │   │ entities     │   │   │
│        │   └─────────────────────────┘   │   └──────────────┘   │   │
│        │                                  │                      │   │
│        ▼                                  │                      │   │
│   ┌─────────────────────────────────┐    │                      │   │
│   │     C++ Native Library          │    │                      │   │
│   │  ┌─────────┬─────────────────┐  │    │                      │   │
│   │  │  SIMD   │  Aho-Corasick   │  │    │                      │   │
│   │  │ Matcher │   Tokenizer     │  │    │                      │   │
│   │  └─────────┴─────────────────┘  │    │                      │   │
│   └─────────────────────────────────┘    │                      │   │
└─────────────────────────────────────────────────────────────────────┘
```

## Language Distribution

| Language | Role | Why |
|----------|------|-----|
| **Go** | Main orchestrator, API server | Native performance, excellent concurrency, fast compilation |
| **Python** | LLM inference, ML pipelines | Rich ML ecosystem, Phi-3 integration |
| **C++** | SIMD pattern matching, tokenization | Maximum CPU performance, AVX2 vectorization |
| **Rust** | Fast classifier, parser | Memory safety, competitive with C++ |

## Features

- **Multi-worker Entity Extraction**: Specialized Phi-3 workers for dates, persons, organizations, amounts
- **Fast Native Extraction**: C++/Go pattern matching (~800ns per operation)
- **Graph Database**: PostgreSQL with graph-like storage for entities and relationships
- **Real-time Chat**: WebSocket support for streaming responses
- **Query Classification**: Intelligent routing to relevant extraction workers
- **Entity Merging**: Automatic deduplication with similarity detection

## Quick Start

### Prerequisites

- Go 1.21+
- Python 3.11+
- PostgreSQL 14+
- CMake 3.16+ (for C++ build)

### 1. Build C++ Library (Optional)

```bash
cd native/cpp
mkdir -p build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
```

### 2. Build Go Server

```bash
# Without CGo (pure Go fallback, recommended for quick start)
CGO_ENABLED=0 go build -o bin/flow-server ./cmd/server

# With CGo (links C++ library for maximum performance)
CGO_ENABLED=1 go build -o bin/flow-server ./cmd/server
```

### 3. Setup Python LLM Service

```bash
cd services/llm
python -m venv venv
source venv/bin/activate
pip install fastapi uvicorn pydantic pydantic-settings httpx python-multipart

# Optional: Install torch for actual Phi-3 inference
# pip install torch transformers accelerate
```

### 4. Setup Database

```bash
# Create database
psql -U postgres -c "CREATE DATABASE flow;"

# Run migrations
psql -U postgres -d flow -f db/migrations/001_init_schema_simple.sql
```

### 5. Run Services

```bash
# Terminal 1: Python LLM Service
cd services/llm
source venv/bin/activate
uvicorn src.main:app --host 0.0.0.0 --port 8001

# Terminal 2: Go Server
PORT=8090 LLM_SERVICE_URL=http://localhost:8001 ./bin/flow-server
```

## API Endpoints

### Health & Status

```bash
curl http://localhost:8090/api/v1/health
```

### Fast Extraction (Native Go - ~800ns)

```bash
# Extract all entity types
curl -X POST http://localhost:8090/api/v1/fast/extract \
  -H "Content-Type: application/json" \
  -d '{"text": "Meeting on 2024-01-15 about $5 million deal with john@acme.com"}'

# Extract specific types
curl -X POST http://localhost:8090/api/v1/fast/extract/dates \
  -H "Content-Type: application/json" \
  -d '{"text": "Deadline is January 20, 2024"}'

# Classify query type
curl -X POST http://localhost:8090/api/v1/fast/classify \
  -H "Content-Type: application/json" \
  -d '{"text": "Who paid the $5 million?"}'
```

### Pipeline Extraction (Go → Python LLM)

```bash
# Extract with all workers
curl -X POST http://localhost:8090/api/v1/pipeline/extract/all \
  -H "Content-Type: application/json" \
  -d '{"text": "CEO John Smith from Acme Corp sent $2.5 million on March 15, 2024"}'

# Extract with specific workers
curl -X POST http://localhost:8090/api/v1/pipeline/extract \
  -H "Content-Type: application/json" \
  -d '{"text": "Meeting on 2024-01-15", "worker_types": ["date"]}'
```

### Graph Operations

```bash
# Store extracted entities
curl -X POST http://localhost:8090/api/v1/graph/store-entities \
  -H "Content-Type: application/json" \
  -d '{
    "entities": [
      {"type": "person", "value": "John Smith", "confidence": 0.9},
      {"type": "organization", "value": "Acme Corp", "confidence": 0.85}
    ],
    "source_text": "John Smith works at Acme Corp"
  }'

# List nodes
curl "http://localhost:8090/api/v1/graph/nodes?label=Person"

# Search nodes
curl "http://localhost:8090/api/v1/graph/search?q=John"

# Get node connections
curl "http://localhost:8090/api/v1/graph/nodes/{node_id}/connections"
```

### Python LLM Service Direct Access

```bash
# Date extraction
curl -X POST http://localhost:8001/extract/date \
  -H "Content-Type: application/json" \
  -d '{"text": "Meeting on 2024-01-15"}'

# Smart routing (auto-selects workers based on query)
curl -X POST http://localhost:8001/extract/smart \
  -H "Content-Type: application/json" \
  -d '{"text": "Who paid $1 million on January 15?"}'
```

## Project Structure

```
flow/
├── cmd/server/              # Go server entry point
├── internal/
│   ├── api/                 # HTTP handlers
│   │   ├── routes_chat.go
│   │   ├── routes_pipeline.go
│   │   ├── routes_fast_extract.go
│   │   └── routes_graph.go
│   ├── config/              # Configuration
│   ├── native/              # CGo bridge to C++
│   │   ├── cgo_bridge.go    # CGo wrapper
│   │   └── pure_go.go       # Pure Go fallback
│   ├── pipeline/            # Go orchestrator
│   └── websocket/           # WebSocket hub
├── native/
│   ├── cpp/                 # C++ SIMD library
│   │   ├── include/
│   │   │   ├── pattern_matcher.hpp
│   │   │   └── pattern_ffi.hpp
│   │   └── src/
│   │       ├── pattern_matcher.cpp
│   │       └── pattern_ffi.cpp
│   └── rust/                # Rust classifier
│       └── src/
│           └── classifier.rs
├── services/
│   ├── llm/                 # Python LLM service
│   │   └── src/
│   │       ├── workers/     # Specialized extractors
│   │       │   ├── date_worker.py
│   │       │   ├── person_worker.py
│   │       │   ├── org_worker.py
│   │       │   └── amount_worker.py
│   │       ├── router/      # Query classification
│   │       └── aggregator/  # Entity merging
│   ├── discord/             # Discord bot
│   └── node/                # Node.js async service
├── db/
│   └── migrations/          # SQL migrations
│       └── 001_init_schema_simple.sql
├── frontend/                # Svelte frontend
└── pkg/                     # Shared Go packages
```

## Performance

| Operation | Engine | Latency |
|-----------|--------|---------|
| Pattern Matching | Native Go | ~800ns |
| Entity Extraction | Native Go | ~55μs |
| LLM Extraction | Python/Phi-3 | ~2-4ms |
| Graph Query | PostgreSQL | ~1-5ms |

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Go server port |
| `LLM_SERVICE_URL` | http://localhost:8001 | Python LLM service URL |
| `PG_HOST` | localhost | PostgreSQL host |
| `PG_PORT` | 5432 | PostgreSQL port |
| `PG_DATABASE` | flow | Database name |
| `PG_USER` | postgres | Database user |
| `PG_PASSWORD` | - | Database password |

## Entity Types

| Type | Worker | Examples |
|------|--------|----------|
| Date | DateWorker | "2024-01-15", "January 20, 2024", "next Tuesday" |
| Person | PersonWorker | "John Smith", "CEO", "john@example.com" |
| Organization | OrgWorker | "Acme Corp", "Goldman Sachs Inc" |
| Amount | AmountWorker | "$5 million", "€500,000", "2 billion USD" |

## Database Schema

```sql
-- Graph nodes (entities as vertices)
CREATE TABLE graph_nodes (
    id UUID PRIMARY KEY,
    label VARCHAR(50),      -- Person, Organization, Amount, Date
    properties JSONB,       -- Entity data
    created_at TIMESTAMP
);

-- Graph edges (relationships)
CREATE TABLE graph_edges (
    id UUID PRIMARY KEY,
    from_node_id UUID,
    to_node_id UUID,
    label VARCHAR(50),      -- WORKS_FOR, PAID, DATED, etc.
    properties JSONB
);

-- Extracted entities with source tracking
CREATE TABLE entities (
    id UUID PRIMARY KEY,
    entity_type VARCHAR(50),
    value TEXT,
    confidence FLOAT,
    source_text TEXT,
    graph_node_id UUID
);
```

## Testing

```bash
# Test Go native extraction
go test -v ./internal/native/...

# Benchmark
go test -bench=. -benchmem ./internal/native/...

# Test Python LLM service
curl http://localhost:8001/health
curl -X POST http://localhost:8001/extract/all \
  -H "Content-Type: application/json" \
  -d '{"text": "Test extraction"}'
```

## License

MIT

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push branch (`git push origin feature/amazing`)
5. Open Pull Request
