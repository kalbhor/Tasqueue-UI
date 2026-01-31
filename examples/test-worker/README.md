# Test Worker Example

This example creates a Tasqueue worker that generates test jobs to demonstrate the Tasqueue-UI dashboard.

## Features

- **Redis broker**: Uses Redis for shared state with the UI
- **Multiple job types**: Add, multiply, and intentional fail operations
- **Chains**: Sequential job execution
- **Groups**: Parallel job execution
- **Retries**: Failed jobs will retry up to 2 times

## Prerequisites

- Redis running on localhost:6379 (or set `REDIS_ADDR` environment variable)
- Go 1.23 or higher

## Usage

### Quick Start

1. Start Redis:
   ```bash
   redis-server
   ```

2. Run the worker:
   ```bash
   cd examples/test-worker
   go run main.go
   ```

The worker will:
1. Register three tasks: `add`, `multiply`, and `fail`
2. Start processing jobs in the background
3. Enqueue 10 individual jobs:
   - 5 add jobs (succeed)
   - 3 multiply jobs (succeed)
   - 2 fail jobs (fail after retries)
4. Enqueue 1 chain of 3 sequential jobs
5. Enqueue 1 group of 4 parallel jobs

## Viewing Jobs in the UI

This worker uses Redis as a shared broker, so jobs will appear in the Tasqueue-UI dashboard in real-time.

### Start Tasqueue-UI

In a separate terminal:

```bash
cd ../..
make build
./bin/tasqueue-ui -broker redis -redis-addr localhost:6379
```

Then open http://localhost:8080 to view the dashboard.

### Using Docker Compose (Easier!)

Instead of running components manually, use Docker Compose to start Redis and the worker:

```bash
# Terminal 1: Start Redis and worker
cd ../
docker-compose up

# Terminal 2: Start UI locally
cd ..
./bin/tasqueue-ui -broker redis
```

This approach uses Docker for infrastructure (Redis + worker) while running the UI locally for easier development.

## Jobs Created

### Individual Jobs
- **Add jobs**: Calculate sum of two numbers (2s processing time)
- **Multiply jobs**: Calculate product of two numbers (1s processing time)
- **Fail jobs**: Always fail for testing error handling (with retries)

### Chain
A sequence of 3 add jobs that execute one after another:
1. 0 + 1 = 1
2. 1 + 1 = 2
3. 2 + 1 = 3

### Group
4 multiply jobs that execute in parallel:
- 0 * 0 = 0
- 2 * 3 = 6
- 4 * 6 = 24
- 6 * 9 = 54

## Configuration

### Redis Address

By default, the worker connects to `127.0.0.1:6379`. To use a different Redis instance:

```bash
REDIS_ADDR=redis.example.com:6379 go run main.go
```

## Customization

You can modify `main.go` to:
- Change the number and types of jobs
- Adjust processing delays
- Add new task types
- Test different failure scenarios
- Experiment with job options (timeouts, schedules, etc.)
- Connect to different Redis instances
