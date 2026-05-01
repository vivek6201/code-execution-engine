# ⚡ Code Execution Engine

A production-grade, distributed code execution engine built in Go. Submit code via REST API, execute it inside isolated Docker containers, and get results — all asynchronously with real-time job status tracking.

Think of it as a self-hosted [Judge0](https://judge0.com/) — designed for coding platforms, online judges, and technical interview systems.

## 🏗️ System Architecture

![System Architecture](docs/architecture.png)

### How It Works

```
Client                API Server           Redis              Workers (x3)         Docker
  │                      │                   │                    │                   │
  ├─POST /api/v1/run────►│                   │                    │                   │
  │                      ├──SET QUEUED───────►│                    │                   │
  │                      ├──LPUSH job────────►│                    │                   │
  │◄──202 {job_id}───────┤                   │                    │                   │
  │                      │                   │◄──BRPOP─────────────┤                   │
  │                      │                   ├──SET PROCESSING────►│                   │
  │                      │                   │                    ├──Create Container──►│
  │                      │                   │                    ├──Copy Code─────────►│
  │                      │                   │                    ├──Compile───────────►│
  │                      │                   │                    ├──Execute (parallel)─►│
  │                      │                   │                    │◄──Results────────────┤
  │                      │                   │                    ├──Remove Container──►│
  │                      │                   │◄──SET result────────┤                   │
  ├─GET /api/v1/run/:id─►│                   │                    │                   │
  │                      ├──GET result───────►│                    │                   │
  │◄──200 {result}───────┤                   │                    │                   │
```

### Execution Flow

1. **Client** submits code + test cases via `POST /api/v1/run`
2. **API Server** generates a job ID, sets status to `QUEUED` in Redis, pushes to job queue
3. **Worker** (one of 3 replicas) pops the job via `BRPOP`, updates status to `PROCESSING`
4. **Worker** creates an isolated Docker container:
   - Pulls the language image (cached after first pull)
   - Creates sandbox container (`sleep infinity`, 256MB RAM, 0.5 CPU, no network)
   - Copies code file via tar archive
   - Compiles once (for compiled languages)
   - Runs all test cases **concurrently** with bounded parallelism (`runtime.NumCPU()`)
   - Cleans up container
5. **Worker** stores final result in Redis
6. **Client** polls `GET /api/v1/run/:id` to get real-time status and results

---

## 🛠️ Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.26 | Core application |
| **Web Framework** | Gin | REST API with CORS support |
| **Message Queue** | Redis 7 (Lists) | Job queue with `LPUSH`/`BRPOP` |
| **Job Status Store** | Redis 7 (KV) | Real-time status tracking with TTL |
| **Container Runtime** | Docker API (Go SDK) | Sandboxed code execution |
| **Orchestration** | Docker Compose | Multi-service deployment |

### Supported Languages

| Language | Image | Compile | Run |
|----------|-------|---------|-----|
| C++ | `gcc:latest` | `g++ main.cpp -o main` | `./main` |
| Python | `python:3.9-slim` | — | `python3 main.py` |
| Java | `openjdk:17.0.1-jdk-slim` | `javac Main.java` | `java Main` |

---

## 🚀 Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development only)

### Run with Docker Compose

```bash
# Clone the repo
git clone https://github.com/vivek6201/code-execution-engine.git
cd code-execution-engine

# Start all services (API + 3 Workers + Redis)
docker compose up --build
```

The API server will be available at `http://localhost:8080`.

### API Endpoints

#### Submit Code

```bash
POST /api/v1/run
Content-Type: application/json

{
  "code": "#include <iostream>\nusing namespace std;\nint main(){\n    int a, b;\n    cin >> a >> b;\n    cout << a + b;\n    return 0;\n}",
  "language": "cpp",
  "test_cases": [
    { "input": "2 3", "expected_output": "5" },
    { "input": "-1 1", "expected_output": "0" },
    { "input": "100 200", "expected_output": "300" }
  ]
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "message": "Job queued successfully",
  "data": {
    "job_id": "8a102565-c43c-464a-bd21-0129435c7e08"
  }
}
```

#### Poll Results

```bash
GET /api/v1/run/:job_id
```

**Response (while processing):**
```json
{
  "success": true,
  "message": "Job is being processed",
  "data": {
    "status": "PROCESSING"
  }
}
```

**Response (completed):**
```json
{
  "success": true,
  "message": "Result fetched successfully",
  "data": {
    "status": "SUCCESS",
    "test_cases": [
      {
        "input": "2 3",
        "expected_output": "5",
        "actual_output": "5",
        "status": "SUCCESS"
      },
      {
        "input": "-1 1",
        "expected_output": "0",
        "actual_output": "0",
        "status": "SUCCESS"
      }
    ],
    "total": 2,
    "passed": 2
  }
}
```

#### Job Status Lifecycle

```
QUEUED → PROCESSING → SUCCESS | FAILED | ERROR | TLE
```

| Status | Description |
|--------|-------------|
| `QUEUED` | Job is waiting in the Redis queue |
| `PROCESSING` | A worker has picked up the job |
| `SUCCESS` | All test cases passed (or single execution succeeded) |
| `FAILED` | One or more test cases produced wrong output |
| `ERROR` | Compilation error or runtime error |
| `TLE` | Time Limit Exceeded (5-second per-test-case timeout) |

---

## 📁 Project Structure

```
code_execution_engine/
├── cmd/
│   ├── api/                    # API server entry point
│   └── worker/                 # Worker entry point
├── config/
│   └── config.go               # Environment-based configuration
├── docker/
│   ├── Dockerfile.api          # API server container
│   └── Dockerfile.worker       # Worker container (with Docker socket)
├── internals/
│   ├── api/                    # HTTP layer
│   │   ├── dtos/               # Request/response DTOs with validation
│   │   ├── handlers/           # Gin route handlers
│   │   ├── routes/             # Route registration
│   │   └── utility/            # Standardized JSON responses
│   ├── app/
│   │   └── bootstrap.go        # Dependency injection & startup
│   ├── core/                   # Business logic (no external deps)
│   │   ├── evaluator/          # Output comparison service
│   │   ├── execution/          # Job orchestration & batch execution
│   │   ├── job/                # Job domain model
│   │   └── result/             # Result types & status enums
│   ├── engine/                 # Execution engine
│   │   ├── executor/           # Language-agnostic executor
│   │   └── runners/            # Runner interface + language impls
│   │       └── languages/      # C++, Python, Java runners
│   └── infra/                  # Infrastructure adapters
│       ├── cache/              # Redis client (shared connection)
│       ├── isolation/          # Docker sandbox (container lifecycle)
│       │   ├── client.go       # Client struct, types, constants
│       │   ├── container.go    # Create, pull image, copy code
│       │   ├── exec.go         # Execute command with timeout
│       │   └── run.go          # Run/RunBatch orchestration
│       └── queue/              # Redis queue (enqueue/dequeue/helpers)
├── docker-compose.yml          # Full stack: Redis + API + 3 Workers
└── go.mod
```

---

## 🔒 Security

| Feature | Implementation |
|---------|---------------|
| **Network Isolation** | `NetworkDisabled: true` — containers cannot make outbound requests |
| **Resource Limits** | 256MB RAM, 0.5 CPU per container |
| **Execution Timeout** | 5s per test case via `timeout -s KILL` + Go-side safety net |
| **No Persistent Storage** | Containers are force-removed after execution |
| **Read-only Code** | Code is copied via tar, not mounted from host |

---

## ⚡ Performance Optimizations

### Batch Execution (Compile Once, Run Many)
Instead of creating a new container per test case, the engine:
1. Creates **1 container** per job
2. Compiles **once**
3. Runs the binary **N times** using `docker exec`

**Result:** 100 test cases in ~2-3s instead of ~18s.

### Concurrent Test Cases
Test cases run in parallel using a bounded goroutine pool (`runtime.NumCPU()` workers). This maximizes throughput while preventing resource exhaustion.

### Early Exit on Failure
When any test case fails (wrong answer, error, or TLE), the engine **cancels remaining test cases** via Go context cancellation. No wasted compute.

### Horizontal Scaling
Workers are stateless — scale by increasing `deploy.replicas` in `docker-compose.yml`:
```yaml
worker:
  deploy:
    replicas: 10  # Scale to 10 workers
```

---

## 🗺️ Future Enhancements

### Authentication & Multi-Tenancy
- [ ] API key management (SHA-256 hashed, Postgres-backed)
- [ ] Role-based access: `client` (execute) and `admin` (manage keys)
- [ ] Per-key rate limiting using Redis sliding window

### Language Support
- [ ] JavaScript/Node.js
- [ ] Rust
- [ ] Go
- [ ] C#

### Execution Features
- [ ] Custom time/memory limits per request
- [ ] Execution time measurement per test case
- [ ] WebSocket support for real-time status updates (replace polling)
- [ ] `--pids-limit` on containers to prevent fork bombs

### Infrastructure
- [ ] OpenTelemetry instrumentation (job duration, failure rates)
- [ ] Prometheus metrics endpoint
- [ ] Dead letter queue for failed jobs
- [ ] Job result webhooks (push instead of poll)
- [ ] Pre-warmed container pools for near-zero cold start

### Security Hardening
- [ ] Seccomp profiles to restrict system calls
- [ ] AppArmor profiles for container isolation
- [ ] Non-root user execution inside containers
- [ ] Read-only filesystem with tmpfs for output

---

## 📊 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection URL |
| `DB_URL` | — | PostgreSQL connection URL (future) |
| `GIN_MODE` | `debug` | Set to `release` for production |

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is open source and available under the [MIT License](LICENSE).
