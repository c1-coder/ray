#!/bin/bash

# Start script for the Ray platform

echo "Starting Ray Platform..."
echo "========================"

# Create bin directory if it doesn't exist
mkdir -p bin

# Build all components
echo "Building components..."
cd /workspace/ray
go build -o bin/master ./cmd/master
go build -o bin/worker ./cmd/worker
go build -o bin/api ./cmd/api
go build -o bin/cli ./cmd/cli

# Check if builds were successful
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"

# Start master node
echo "Starting master node..."
cd /workspace/ray
./bin/master &
MASTER_PID=$!
sleep 5

# Start API server
echo "Starting API server..."
cd /workspace/ray
./bin/api &
API_PID=$!
sleep 3

# Start worker nodes
echo "Starting worker nodes..."
for i in {1..3}
do
    cd /workspace/ray
    WORKER_ADDRESS="localhost" MASTER_ADDRESS="localhost:50051" ./bin/worker &
    WORKER_PIDS[$i]=$!
    sleep 2
done

echo ""
echo "Ray Platform started successfully!"
echo "================================"
echo "Master node: localhost:50051"
echo "API server: http://localhost:8080"
echo "Web UI: http://localhost:8080 (when frontend is running)"
echo ""
echo "To access the CLI:"
echo "  cd /workspace/ray && ./bin/cli"
echo ""
echo "To stop the platform, run:"
echo "  pkill -f 'bin/master|bin/worker|bin/api'"
echo ""
echo "Waiting for platform to stabilize..."
sleep 5

# Show initial status
echo ""
echo "Initial status:"
echo "=============="
cd /workspace/ray
./bin/cli -cmd status
