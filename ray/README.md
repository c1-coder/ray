# Distributed Task Management Platform

A powerful distributed task management platform built to coordinate and manage tasks across multiple servers. This platform allows you to bind multiple servers together, issue tasks, and have them autonomously allocate and process tasks with built-in fault tolerance.

## Features

- **Distributed Task Management**: Coordinate tasks across multiple servers
- **Autonomous Task Allocation**: Servers can autonomously allocate tasks based on their capabilities and load
- **Fault Tolerance**: If a server goes down, other servers can take over its tasks
- **Security Management**: API key management and input sanitization
- **Vulnerability Scanning**: Scan integrated tools for vulnerabilities
- **Real-time Monitoring**: Monitor server health and task progress

## Architecture

The platform consists of two main components:

1. **Master Node**: Coordinates task distribution and manages worker nodes
2. **Worker Nodes**: Execute tasks and report back results

## Getting Started

### Prerequisites

- Go 1.20 or later
- gRPC and Protocol Buffers

### Installation

1. Clone the repository

```bash
git clone https://github.com/yourusername/ray.git
cd ray
```

2. Install dependencies

```bash
go mod download
go mod tidy
```

3. Build the platform

```bash
go build -o bin/master ./cmd/master
go build -o bin/worker ./cmd/worker
```

### Usage

#### Starting the Master Node

```bash
./bin/master
```

The master node will start on port 50051 by default.

#### Starting Worker Nodes

```bash
WORKER_ADDRESS="localhost" MASTER_ADDRESS="localhost:50051" ./bin/worker
```

You can start multiple worker nodes on different servers by setting the appropriate environment variables.

### Task Management

#### Creating a Task

You can create tasks using the gRPC API or by using the provided test script:

```bash
./test_platform.sh
```

#### Task Types

- **Monitoring**: Task for monitoring system resources
- **AI**: Task for AI-related processing
- **Communication**: Task for communication-related operations
- **Custom**: Custom task with user-defined payload

### Fault Tolerance

The platform automatically handles worker node failures:

1. When a worker node goes down, the master node detects it through heartbeat timeout
2. The master node reassigns the failed node's tasks to other available worker nodes
3. The platform continues to process tasks without interruption

### Security

The platform includes security features:

- API key management for authentication
- Input sanitization to prevent injection attacks
- Vulnerability scanning for integrated tools

### Integration with External Tools

The platform can integrate with external tools like:

- **Nezha**: Server monitoring tool
- **Hermes Agent**: AI agent framework
- **OpenClaw**: Web crawling tool

## Configuration

### Environment Variables

- **MASTER_ADDRESS**: Address of the master node (e.g., "localhost:50051")
- **WORKER_ADDRESS**: Address of the worker node (e.g., "localhost")

### Master Node Configuration

- **Heartbeat Timeout**: 30 seconds by default
- **Check Interval**: 10 seconds by default

### Worker Node Configuration

- **Heartbeat Interval**: 5 seconds by default
- **Task Polling Interval**: 2 seconds by default

## Troubleshooting

### Common Issues

1. **Port Already in Use**: Make sure no other process is using port 50051
2. **Worker Node Registration Failed**: Check the master address and network connectivity
3. **Task Processing Failed**: Check the worker node logs for error messages

### Logs

- Master node logs: Output to stdout
- Worker node logs: Output to stdout

## Scaling

To scale the platform:

1. Add more worker nodes
2. Distribute worker nodes across different servers
3. Configure load balancing if needed

## License

MIT
