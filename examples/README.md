# Tasqueue-UI Examples

This directory contains examples and tools for testing and demonstrating Tasqueue-UI.

## Quick Start with Docker Compose

The easiest way to set up the test environment:

```bash
cd examples
docker-compose up
```

This starts:
- **Redis** on port 6379 (shared broker/results store)
- **Test Worker** (automatically creates and processes jobs)

Then run the UI locally:

```bash
cd ..
make build
./bin/tasqueue-ui -broker redis -redis-addr localhost:6379
```

Open http://localhost:8080 to view the dashboard!

### What You'll See

The test worker automatically creates:
- 5 add jobs (will succeed after ~2s each)
- 3 multiply jobs (will succeed after ~1s each)
- 2 fail jobs (will fail after retries)
- 1 chain of 3 sequential add jobs
- 1 group of 4 parallel multiply jobs

All jobs are visible in the UI dashboard in real-time.

### Stopping the Services

```bash
docker-compose down
```

To also remove the Redis data volume:

```bash
docker-compose down -v
```

## Manual Setup (Without Docker)

If you prefer to run everything manually:

### 1. Start Redis

```bash
redis-server
```

### 2. Run Test Worker

```bash
cd test-worker
go run main.go
```

### 3. Start Tasqueue-UI

```bash
cd ..
make build
./bin/tasqueue-ui -broker redis -redis-addr localhost:6379
```

Open http://localhost:8080 to view the dashboard.

## Test Worker

The `test-worker` directory contains a standalone worker application that:
- Creates various types of jobs (add, multiply, fail)
- Demonstrates chains and groups
- Shows retry behavior with intentional failures
- Uses Redis broker by default (configurable via `REDIS_ADDR` env var)

See [test-worker/README.md](test-worker/README.md) for more details.

## Architecture

```
                    ┌──────────────────┐
                    │  Docker Compose  │
                    │                  │
┌──────────────┐    │  ┌───────────┐  │    ┌──────────────┐
│              │    │  │   Redis   │  │    │              │
│ Tasqueue-UI  │◀───┼──│  :6379    │◀─┼────│ Test Worker  │
│   (local)    │    │  └───────────┘  │    │  (container) │
│              │    │                  │    │              │
└──────────────┘    └──────────────────┘    └──────────────┘
```

- **Test Worker (Docker)**: Registers tasks, processes jobs, and enqueues new jobs
- **Redis (Docker)**: Shared message broker and results store
- **Tasqueue-UI (Local)**: Read-only monitoring dashboard running on your machine

## Environment Variables

### Test Worker
- `REDIS_ADDR`: Redis server address (default: `localhost:6379`, set to `redis:6379` in docker-compose)

## Troubleshooting

### Jobs not appearing in UI

1. Ensure all components are using the same Redis instance
2. Check that Redis is running: `redis-cli ping`
3. Verify connection in logs

### Port already in use

If port 8080 is already in use, modify `docker-compose.yml`:

```yaml
tasqueue-ui:
  ports:
    - "3000:8080"  # Use port 3000 instead
```

### Docker build issues

Ensure you're running docker-compose from the `examples` directory and that the parent Tasqueue-UI directory structure is intact.
