# ⚡ Code Execution Engine

A production-grade, distributed code execution engine built in Go. Submit code via REST API, execute it inside isolated Docker containers, and get results — all asynchronously with real-time job status tracking.

Think of it as a self-hosted [Judge0](https://judge0.com/) — designed for coding platforms, online judges, and technical interview systems. It now includes full SaaS features like JWT authentication, Tiered Subscription Plans, and API Key management!

## 🏗️ System Architecture

![System Architecture](docs/architecture.png)

### Execution Flow

1. **Client** submits code + test cases via `POST /api/v1/judge/run`
2. **API Server** generates a job ID, sets status to `QUEUED` in Redis, pushes to the job queue
3. **Worker** (one of 3 replicas) pops the job via `BRPOP`, updates status to `PROCESSING`
4. **Worker** creates a highly-isolated Docker container:
   - Creates sandbox container (`sleep infinity`, 256MB RAM, 0.5 CPU, no network, non-root user `1000:1000`, dropped capabilities)
   - Copies code file via memory
   - Compiles once (for compiled languages)
   - Runs all test cases **concurrently** with bounded parallelism (`runtime.NumCPU()`)
   - Measures exact `time_ms` per test case and fetches total peak `memory_kb` via Linux cgroups
   - Cleans up container
5. **Worker** stores final result in Redis
6. **Client** polls `GET /api/v1/judge/run/:id` to get real-time status and execution metrics

---

## 🛠️ Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.26 | Core application |
| **Web Framework** | Gin | REST API with CORS support |
| **Database** | PostgreSQL + GORM | User accounts, Subscriptions, API Keys |
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
| Node.js | `node:22.18-alpine` | — | `node main.js` |

---

## 🚀 Getting Started

### Prerequisites

- Docker & Docker Desktop

### Run with Docker Compose

```bash
# Clone the repo
git clone https://github.com/vivek6201/code-execution-engine.git
cd code-execution-engine

# Start all services (API + Postgres + 3 Workers + Redis)
docker compose up --build
```

The API server will be available at `http://localhost:8080`.

---

## 🔒 Security Hardening

Running untrusted code requires extreme paranoia. We use the following layers of isolation:

| Feature | Implementation |
|---------|---------------|
| **Network Isolation** | `NetworkDisabled: true` — containers cannot make outbound requests |
| **Resource Limits** | 256MB RAM, 0.5 CPU per container |
| **Process Limits** | `PidsLimit: 64` prevents fork bombs and massive threading |
| **Privilege Escalation** | `SecurityOpt: ["no-new-privileges:true"]`, `CapDrop: ["ALL"]` |
| **Execution Timeout** | 5s per test case via `timeout -s KILL` + Go-side safety net |
| **No Persistent Storage** | Containers are force-removed immediately after execution |

---

## ⚡ Performance Optimizations

### Batch Execution (Compile Once, Run Many)
Instead of creating a new container per test case, the engine:
1. Creates **1 container** per job
2. Compiles **once**
3. Runs the binary **N times** concurrently using `docker exec`

### Exact Execution Metrics (Time & Memory)
The engine returns high-precision metrics directly from the kernel:
- **`time_ms`**: Measured via Go's internal high-precision stopwatch wrapping the `docker exec` call for each test case.
- **`memory_kb`**: The absolute peak memory consumed by the entire container, read directly from the Linux kernel's cgroups (`/sys/fs/cgroup/memory.peak` or `memory.max_usage_in_bytes`).

### Early Exit on Failure
When any test case fails (wrong answer, error, or TLE), the engine **cancels remaining test cases** via Go context cancellation. No wasted compute.

---

## 🔐 API Examples

### Submit Code

```bash
POST /api/v1/judge/run
Content-Type: application/json

{
  "code": "#include <iostream>\nusing namespace std;\nint main(){\n    int a, b;\n    cin >> a >> b;\n    cout << a + b;\n    return 0;\n}",
  "language": "cpp",
  "test_cases": [
    { "input": "2 3", "expected_output": "5" },
    { "input": "-1 1", "expected_output": "0" }
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

### Poll Results

```bash
GET /api/v1/judge/run/8a102565-c43c-464a-bd21-0129435c7e08
```

**Response (completed):**
```json
{
  "success": true,
  "message": "Result fetched successfully",
  "data": {
    "status": "SUCCESS",
    "time_ms": 142,
    "memory_kb": 4210,
    "test_cases": [
      {
        "input": "2 3",
        "expected_output": "5",
        "actual_output": "5",
        "status": "SUCCESS",
        "time_ms": 142
      }
    ],
    "total": 2,
    "passed": 2
  }
}
```

---

## 🗺️ Future Enhancements

### Execution Features
- [ ] WebSocket support for real-time status updates (replace REST polling)
- [ ] Language Support: Go, Rust, C#

### Infrastructure
- [ ] **Transition to Firecracker MicroVMs or gVisor** for hardware-level isolation.
- [ ] OpenTelemetry instrumentation (job duration, failure rates)
- [ ] Dead letter queue for failed/poison-pill jobs
- [ ] Pre-warmed container pools for near-zero cold starts

---

## 📄 License

This project is open source and available under the [MIT License](LICENSE).
