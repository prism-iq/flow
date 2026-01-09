# Flow Project System Prompt

## Project Overview

Flow is a high-performance polyglot web chatbot system designed for entity extraction and graph-based knowledge management. The architecture leverages multiple programming languages, each chosen for their specific strengths.

## Architecture Philosophy

### Language Distribution

| Language | Role | Why |
|----------|------|-----|
| **Go** | Main orchestrator, API server, concurrency | Native performance, excellent concurrency with goroutines, strong typing, fast compilation |
| **C++** | SIMD pattern matching, tokenization, memory management | Maximum performance for CPU-bound operations, AVX2/AVX512 vectorization, zero-overhead abstractions |
| **Rust** | Fast classifier, parser, FFI safety | Memory safety without GC, excellent FFI, competitive with C++ performance |
| **Python** | LLM inference, ML pipelines | Rich ML ecosystem, Phi-3 integration, rapid prototyping |
| **Node.js** | Async I/O, streaming, Discord bot | Event-driven architecture, excellent for I/O-bound tasks |

### Core Components

```
┌─────────────────────────────────────────────────────────────────┐
│                        Flow Architecture                         │
├─────────────────────────────────────────────────────────────────┤
│  Frontend (Svelte)  │  Discord Bot (Node.js)  │  REST API (Go)  │
├─────────────────────────────────────────────────────────────────┤
│                    Go Orchestrator                               │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐     │
│  │ HTTP Router │ WebSocket   │ Pipeline    │ Graph Svc   │     │
│  │   (Gin)     │    Hub      │ Orchestrator│             │     │
│  └─────────────┴─────────────┴─────────────┴─────────────┘     │
├─────────────────────────────────────────────────────────────────┤
│                    Native Layer (CGo)                            │
│  ┌─────────────────────────────┬────────────────────────────┐  │
│  │   C++ (libflow_synapses)    │    Rust (libflow_parser)   │  │
│  │   - SIMD Pattern Matcher    │    - Fast Classifier       │  │
│  │   - Aho-Corasick            │    - Text Chunker          │  │
│  │   - Fast Tokenizer          │    - Tokenizer             │  │
│  │   - Entity Matcher          │                            │  │
│  └─────────────────────────────┴────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                    Python LLM Service                            │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐     │
│  │ DateWorker  │PersonWorker │ OrgWorker   │AmountWorker │     │
│  │   (Phi-3)   │   (Phi-3)   │   (Phi-3)   │   (Phi-3)   │     │
│  └─────────────┴─────────────┴─────────────┴─────────────┘     │
│  ┌─────────────┬─────────────┬─────────────────────────────┐   │
│  │ QueryRouter │ Classifier  │     Haiku Aggregator        │   │
│  └─────────────┴─────────────┴─────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                    Database Layer                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │   PostgreSQL + Apache AGE (Graph Extension)              │  │
│  │   - Entities as Nodes                                     │  │
│  │   - Relationships as Edges                                │  │
│  │   - Cypher Queries for Graph Traversal                    │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Current State

### Completed Components

1. **Go Server** (`cmd/server/main.go`)
   - Gin HTTP router with CORS
   - WebSocket hub for real-time chat
   - Pipeline orchestrator for concurrent worker execution
   - API routes: `/api/chat`, `/api/pipeline/*`

2. **C++ Native Library** (`native/cpp/`)
   - `SIMDMatcher`: AVX2-accelerated pattern matching
   - `AhoCorasick`: Multi-pattern string matching automaton
   - `FastTokenizer`: High-speed text tokenization
   - `EntityMatcher`: Combined entity extraction
   - FFI exports for CGo integration

3. **Rust Native Library** (`native/rust/`)
   - `FastClassifier`: Query classification with regex
   - `TextChunker`: Text segmentation
   - `Tokenizer`: Rust-native tokenization

4. **Python LLM Service** (`services/llm/`)
   - Specialized Phi-3 workers: Date, Person, Org, Amount
   - Query router with classification
   - Haiku aggregator for validation
   - Entity merger for deduplication
   - FastAPI endpoints: `/extract/{type}`, `/extract/all`, `/extract/smart`

5. **Database Schema** (`db/migrations/`)
   - PostgreSQL tables for conversations, messages
   - Apache AGE graph for entities and relationships
   - Graph operations for node/edge creation

6. **Go CGo Bridge** (`internal/native/`)
   - CGo wrapper for C++ library
   - Pure Go fallback when CGo unavailable

### Pipeline Flow

```
1. User Query
      │
      ▼
2. Fast Classification (Rust/C++)
   │ Keyword + Pattern scoring
   │ Determines relevant worker types
      │
      ▼
3. Parallel Worker Dispatch (Go)
   │ Concurrent goroutines
   │ Context with timeout
      │
      ▼
4. Specialized Extraction (Python/Phi-3)
   │ DateWorker → dates, deadlines
   │ PersonWorker → names, emails
   │ OrgWorker → companies, institutions
   │ AmountWorker → prices, quantities
      │
      ▼
5. Entity Merging (Python)
   │ Deduplication
   │ Confidence scoring
      │
      ▼
6. Haiku Validation (Python)
   │ Cross-reference entities
   │ Generate graph operations
      │
      ▼
7. Graph Storage (Go → PostgreSQL/AGE)
   │ Create nodes for entities
   │ Create edges for relationships
      │
      ▼
8. Response to User
```

## Next Steps

### Immediate

1. **Integration Test**: Build and test the full pipeline end-to-end
2. **CGo Build**: Verify Go can link against C++ library
3. **Performance Benchmark**: Compare C++ vs Pure Go pattern matching

### Short-term

1. **LLM Integration**: Connect Python workers to actual Phi-3 model
2. **Graph Queries**: Implement Cypher query builder in Go
3. **Frontend**: Complete Svelte chat interface
4. **Discord Bot**: Implement /chat and /search commands

### Medium-term

1. **Caching Layer**: Redis for entity cache
2. **Batch Processing**: Async job queue for large documents
3. **Multi-tenant**: User isolation and rate limiting
4. **Monitoring**: Prometheus metrics, Grafana dashboards

## Build Commands

```bash
# Build C++ library
cd native/cpp && mkdir -p build && cd build && cmake .. && make -j

# Build Rust library (requires cargo)
cd native/rust && cargo build --release

# Build Go server (with CGo)
CGO_ENABLED=1 go build -o bin/flow-server ./cmd/server

# Build Go server (pure Go fallback)
CGO_ENABLED=0 go build -o bin/flow-server ./cmd/server

# Run Python LLM service
cd services/llm && uvicorn src.main:app --host 0.0.0.0 --port 8001
```

## Design Principles

1. **Performance First**: Use C++ for hot paths, Go for concurrency, Rust for safety
2. **Language Agnostic**: Each component can be replaced independently
3. **Graceful Degradation**: Pure Go fallback when native libs unavailable
4. **Graph-Native**: Entities stored as nodes, relationships as edges
5. **Parallel by Default**: Concurrent extraction workers, async I/O
