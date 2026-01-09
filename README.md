# Flow

A polyglot AI chatbot platform with high-performance native modules and parallel LLM inference.

## Architecture

```
                    +------------------+
                    |   Svelte UI      |
                    |   (Frontend)     |
                    +--------+---------+
                             |
                    +--------v---------+
                    |  Go Orchestrator |
                    |   (API Gateway)  |
                    +--------+---------+
                             |
          +------------------+------------------+
          |                  |                  |
+---------v------+  +--------v--------+  +------v--------+
| Node.js Async  |  | Python LLM      |  | PostgreSQL    |
| I/O Service    |  | (Phi-3 Workers) |  | + Apache AGE  |
+----------------+  +-----------------+  +---------------+
                             |
          +------------------+------------------+
          |                                     |
+---------v---------+              +-----------v----------+
|   Rust Parser     |              |    C++ Synapses      |
| (High-perf text)  |              |  (SIMD Tensors/FFI)  |
+-------------------+              +----------------------+
```

## Components

| Component | Language | Purpose |
|-----------|----------|---------|
| Orchestrator | Go | API gateway, routing, WebSocket management |
| LLM Service | Python | Parallel Phi-3 inference with Haiku validation |
| Async I/O | Node.js | Task queues, streaming, proxy |
| Parser | Rust | High-performance text parsing, chunking |
| Synapses | C++ | SIMD tensor operations, FFI bridge |
| Frontend | Svelte | Modern chat UI |
| Discord Bot | Node.js | Discord interface |

## Requirements

- Go 1.21+
- Node.js 20+
- Python 3.11+
- Rust 1.75+ (optional)
- CMake 3.16+ (optional)
- PostgreSQL 16+
- Redis 7+

## Quick Start

```bash
# Install dependencies
./scripts/install.sh

# Build all components
./scripts/build.sh

# Start services
./scripts/start.sh
```

## Development

```bash
# Start dev environment with hot reload
./scripts/dev.sh
```

Access:
- **API**: http://localhost:8080
- **Node Service**: http://localhost:3001
- **Frontend**: http://localhost:5173

## Docker

```bash
# Build all images
docker compose build

# Start stack
docker compose up -d

# View logs
docker compose logs -f
```

## Configuration

Copy and edit the example config:

```bash
cp configs/config.example.yaml configs/config.yaml
cp .env.example .env
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Go server port |
| `LOG_LEVEL` | info | Log level |
| `LLM_SERVICE_URL` | http://localhost:8001 | LLM service URL |
| `LLM_NUM_WORKERS` | 4 | Parallel LLM workers |
| `LLM_MODEL_NAME` | microsoft/Phi-3-mini-4k-instruct | Model to use |
| `PG_HOST` | localhost | PostgreSQL host |
| `DISCORD_TOKEN` | - | Discord bot token |

## API Endpoints

### Chat

```bash
# Send message
curl -X POST http://localhost:8080/api/v1/chat/send \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello!", "use_rag": false}'

# Stream response
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "Explain quantum computing"}'
```

### WebSocket

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.send(JSON.stringify({
  type: 'chat',
  content: 'Hello!',
  useRag: false
}));
```

### Health

```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/status
```

## Discord Bot

```bash
# Set environment
export DISCORD_TOKEN=your_token
export DISCORD_CLIENT_ID=your_client_id

# Deploy commands
cd services/discord
npm run deploy

# Start bot
npm start
```

Commands:
- `/chat <message>` - Chat with Flow AI
- `/search <query>` - Search knowledge base
- `/clear` - Clear conversation
- `/status` - Service status

## Native Modules

### Rust Parser

```bash
cd native/rust
cargo build --release
```

Features:
- Markdown parsing
- Text chunking with overlap
- Parallel tokenization
- FFI bindings

### C++ Synapses

```bash
cd native/cpp
mkdir build && cd build
cmake -DCMAKE_BUILD_TYPE=Release ..
make -j$(nproc)
```

Features:
- SIMD tensor operations (AVX2)
- Memory pool allocation
- Neural network layers
- FFI bridge for Go/Rust

## Testing

```bash
# Run all tests
./scripts/test.sh

# Specific language
./scripts/test.sh go
./scripts/test.sh rust
./scripts/test.sh python
```

## Performance

Benchmarks on commodity hardware (Ryzen 5, 32GB RAM, RTX 3060):

| Metric | Value |
|--------|-------|
| API throughput | 443 req/s |
| LLM latency (cold) | ~2s |
| LLM latency (warm) | ~500ms |
| WebSocket connections | 10k+ |
| Memory (idle) | ~200MB |

## Project Structure

```
flow/
├── cmd/server/          # Go entrypoint
├── internal/            # Go internal packages
│   ├── api/            # HTTP handlers
│   ├── config/         # Configuration
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   ├── services/       # Business logic
│   └── websocket/      # WebSocket handling
├── pkg/                 # Go shared packages
├── services/
│   ├── llm/            # Python LLM service
│   ├── node/           # Node.js async service
│   └── discord/        # Discord bot
├── native/
│   ├── rust/           # Rust parser
│   └── cpp/            # C++ synapses
├── frontend/           # Svelte UI
├── scripts/            # Build/deploy scripts
└── configs/            # Configuration files
```

## License

MIT

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push branch (`git push origin feature/amazing`)
5. Open Pull Request
