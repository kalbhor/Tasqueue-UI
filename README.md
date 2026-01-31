# Tasqueue UI

A modern web-based dashboard for monitoring and managing [Tasqueue](https://github.com/kalbhor/tasqueue) job queues.

## Features

- **Dashboard Overview**: Real-time statistics on pending, successful, and failed jobs
- **Job Management**: View, filter, and monitor individual jobs
- **Chain Visualization**: Track sequential job execution with visual progress indicators
- **Group Monitoring**: Monitor parallel job groups and their completion status
- **Multiple Broker Support**: Works with Redis, NATS JetStream, and in-memory brokers
- **Clean UI**: Simple, responsive interface built with vanilla JavaScript
- **Single Binary**: Embedded assets, easy deployment

## Screenshots

### Dashboard
View overview statistics, registered tasks, and queue information.

### Jobs View
Filter and inspect individual jobs with detailed information including payloads, results, and retry status.

### Chains & Groups
Visualize job chains and monitor group execution status.

## Installation

### From Source

```bash
git clone https://github.com/kalbhor/tasqueue-ui.git
cd tasqueue-ui
make build
```

The binary will be available at `bin/tasqueue-ui`.

### Using Go Install

```bash
go install github.com/kalbhor/tasqueue-ui/cmd/server@latest
```

## Quick Start with Docker Compose

The easiest way to test Tasqueue-UI:

```bash
# Start Redis and test worker
cd examples
docker-compose up

# In another terminal, run the UI
cd ..
make build
./bin/tasqueue-ui -broker redis
```

This starts:
- Redis (message broker in Docker)
- Test worker (generates sample jobs in Docker)
- Tasqueue-UI dashboard on http://localhost:8080 (local)

Open http://localhost:8080 to see the dashboard with live job data!

## Usage

### Basic Usage

Start the UI server with default settings (Redis on localhost:6379):

```bash
./bin/tasqueue-ui
```

The dashboard will be available at `http://localhost:8080`.

### Command Line Options

```bash
./bin/tasqueue-ui [options]

Options:
  -port string
        HTTP server port (default "8080")
  -host string
        HTTP server host (default "0.0.0.0")
  -broker string
        Broker type: redis, nats-js, or in-memory (default "redis")
  -redis-addr string
        Redis address (default "localhost:6379")
  -redis-pass string
        Redis password (default "")
  -redis-db int
        Redis database number (default 0)
  -version
        Show version information
```

### Examples

**Using Redis:**
```bash
./bin/tasqueue-ui -broker redis -redis-addr localhost:6379
```

**Using In-Memory Broker (for development/testing):**
```bash
./bin/tasqueue-ui -broker in-memory
```

**Custom Port:**
```bash
./bin/tasqueue-ui -port 3000
```

## Development

### Prerequisites

- Go 1.23 or higher
- Redis (for Redis broker) or NATS (for NATS broker)

### Running in Development Mode

```bash
make dev
```

This starts the server with an in-memory broker for quick testing.

### Project Structure

```
tasqueue-ui/
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/             # HTTP handlers and routes
│   ├── service/         # Tasqueue service layer
│   └── config/          # Configuration management
├── web/
│   ├── static/          # CSS and JavaScript files
│   │   ├── css/
│   │   └── js/
│   └── templates/       # HTML templates
├── go.mod
├── Makefile
└── README.md
```

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Other Commands

```bash
make help              # Show all available commands
make clean             # Remove build artifacts
make fmt               # Format code
make vet               # Run go vet
make install           # Install to GOPATH/bin
```

### Test Worker Example

The `examples/test-worker` directory contains a sample worker that generates test jobs to demonstrate the UI.

**Using Docker Compose (Recommended)**:
```bash
# Terminal 1: Start Redis and worker
cd examples
docker-compose up

# Terminal 2: Start UI
./bin/tasqueue-ui -broker redis
```

**Manual setup** (everything local):
```bash
# Terminal 1: Redis
redis-server

# Terminal 2: Test worker
cd examples/test-worker
go run main.go

# Terminal 3: UI
./bin/tasqueue-ui -broker redis
```

The worker creates:
- Various job types: add, multiply, and intentional failures
- Chains (sequential jobs) and groups (parallel jobs)
- Demonstrates retry behavior with failed jobs

See `examples/README.md` and `examples/test-worker/README.md` for more details.

## API Endpoints

The server exposes the following REST API endpoints:

### Dashboard
- `GET /api/stats` - Dashboard statistics

### Jobs
- `GET /api/jobs?status={status}` - List jobs by status (successful, failed)
- `GET /api/jobs/{id}` - Get specific job details
- `GET /api/jobs/pending/{queue}` - Get pending jobs for a queue
- `DELETE /api/jobs/{id}` - Delete job metadata

### Chains
- `GET /api/chains` - List chains (placeholder)
- `GET /api/chains/{id}` - Get chain details

### Groups
- `GET /api/groups` - List groups (placeholder)
- `GET /api/groups/{id}` - Get group details

### Health
- `GET /health` - Health check endpoint

## Configuration

The UI server connects to the same broker and results backend that your Tasqueue workers use. Make sure to configure the correct broker type and connection details.

### Connecting to Tasqueue

Tasqueue UI is a read-only monitoring tool. It:
- Does **NOT** register task handlers
- Does **NOT** process jobs
- Only **reads** job metadata and status from the results store

Your actual job workers should continue running separately with registered task handlers.

## How It Works

1. **Service Layer**: Initializes a Tasqueue server instance with the specified broker and results backend
2. **API Layer**: Exposes REST endpoints to query job data
3. **Frontend**: Single-page application that polls the API for updates
4. **Auto-Refresh**: Dashboard auto-refreshes every 3 seconds

## Limitations

- **Listing Chains/Groups**: Currently requires manual ID input. Full listing would require scanning the results store (Redis SCAN operation).
- **Read-Only**: UI cannot enqueue new jobs or modify existing ones.
- **Broker Support**: NATS JetStream support is planned but not yet implemented.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the same license as Tasqueue.

## Acknowledgments

Built with [Tasqueue](https://github.com/kalbhor/tasqueue) - A simple, lightweight distributed job/worker implementation in Go.

## Roadmap

- [ ] NATS JetStream broker support
- [ ] WebSocket support for real-time updates
- [ ] Job enqueue interface
- [ ] Advanced filtering and search
- [ ] Job retry/requeue functionality
- [ ] Dark mode
- [ ] Export job data (CSV/JSON)
- [ ] Pagination for large job lists
