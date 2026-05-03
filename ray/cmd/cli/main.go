package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	pb "ray/src/common"
)

var (
	masterAddr = flag.String("master", "localhost:50051", "Master node address")
	command    = flag.String("cmd", "", "Command to execute")
)

func main() {
	flag.Parse()

	// Connect to the master node
	conn, err := grpc.Dial(*masterAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to master: %v", err)
	}
	defer conn.Close()

	// Create clients
	taskClient := pb.NewTaskServiceClient(conn)
	registryClient := pb.NewServiceRegistryServiceClient(conn)

	// Execute command
	if *command != "" {
		executeCommand(*command, taskClient, registryClient)
		return
	}

	// Interactive mode
	interactiveMode(taskClient, registryClient)
}

// executeCommand executes a single command
func executeCommand(cmd string, taskClient pb.TaskServiceClient, registryClient pb.ServiceRegistryServiceClient) {
	switch strings.ToLower(cmd) {
	case "status":
		printStatus(taskClient, registryClient)
	case "tasks":
		printTasks(taskClient)
	case "nodes":
		printNodes(registryClient)
	case "create-task":
		createTask(taskClient)
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printHelp()
	}
}

// interactiveMode starts interactive mode
func interactiveMode(taskClient pb.TaskServiceClient, registryClient pb.ServiceRegistryServiceClient) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Ray Cluster CLI - Type 'help' for commands")
	fmt.Println("========================================")

	for {
		fmt.Print("ray> ")
		if !scanner.Scan() {
			break
		}

		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			continue
		}

		switch strings.ToLower(cmd) {
		case "status":
			printStatus(taskClient, registryClient)
		case "tasks":
			printTasks(taskClient)
		case "nodes":
			printNodes(registryClient)
		case "create-task":
			createTask(taskClient)
		case "help":
			printHelp()
		case "exit", "quit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			printHelp()
		}
	}
}

// printStatus prints the current status of the cluster
func printStatus(taskClient pb.TaskServiceClient, registryClient pb.ServiceRegistryServiceClient) {
	fmt.Println("Cluster Status")
	fmt.Println("==============")

	// Get nodes
	nodesResp, err := registryClient.GetAllNodes(context.Background(), &pb.GetAllNodesRequest{})
	if err != nil {
		fmt.Printf("Error getting nodes: %v\n", err)
		return
	}

	fmt.Printf("Nodes: %d\n", len(nodesResp.Nodes))
	for _, node := range nodesResp.Nodes {
		fmt.Printf("  - %s (%s): %s\n", node.Id, node.Address, node.Status)
	}

	// Get tasks
	tasksResp, err := taskClient.GetPendingTasks(context.Background(), &pb.GetPendingTasksRequest{})
	if err != nil {
		fmt.Printf("Error getting tasks: %v\n", err)
		return
	}

	fmt.Printf("Pending Tasks: %d\n", len(tasksResp.Tasks))

	fmt.Println()
}

// printTasks prints all tasks
func printTasks(taskClient pb.TaskServiceClient) {
	fmt.Println("Tasks")
	fmt.Println("=====")

	// Get pending tasks
	pendingResp, err := taskClient.GetPendingTasks(context.Background(), &pb.GetPendingTasksRequest{})
	if err != nil {
		fmt.Printf("Error getting pending tasks: %v\n", err)
		return
	}

	fmt.Println("Pending Tasks:")
	for _, task := range pendingResp.Tasks {
		fmt.Printf("  - %s: %s (Priority: %s)\n", task.Id, task.Type, task.Priority)
	}

	fmt.Println()
}

// printNodes prints all nodes
func printNodes(registryClient pb.ServiceRegistryServiceClient) {
	fmt.Println("Nodes")
	fmt.Println("=====")

	nodesResp, err := registryClient.GetAllNodes(context.Background(), &pb.GetAllNodesRequest{})
	if err != nil {
		fmt.Printf("Error getting nodes: %v\n", err)
		return
	}

	for _, node := range nodesResp.Nodes {
		fmt.Printf("- %s\n", node.Id)
		fmt.Printf("  Address: %s\n", node.Address)
		fmt.Printf("  Status: %s\n", node.Status)
		fmt.Printf("  Type: %s\n", node.Type)
		fmt.Printf("  Capabilities: %v\n", node.Capabilities)
		fmt.Println()
	}
}

// createTask creates a new task
func createTask(taskClient pb.TaskServiceClient) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Create New Task")
	fmt.Println("==============")

	// Get task type
	fmt.Print("Task Type (monitoring/ai/communication/custom): ")
	scanner.Scan()
	taskType := strings.TrimSpace(scanner.Text())

	// Get priority
	fmt.Print("Priority (low/medium/high/critical): ")
	scanner.Scan()
	priority := strings.TrimSpace(scanner.Text())

	// Get payload
	fmt.Print("Payload (JSON): ")
	scanner.Scan()
	payloadStr := strings.TrimSpace(scanner.Text())

	// Parse payload
	var payloadMap map[string]interface{}
	if payloadStr != "" {
		if err := json.Unmarshal([]byte(payloadStr), &payloadMap); err != nil {
			fmt.Printf("Invalid JSON: %v\n", err)
			return
		}
	}

	// Convert payload to structpb.Struct
	payload, err := structpb.NewStruct(payloadMap)
	if err != nil {
		fmt.Printf("Error creating payload: %v\n", err)
		return
	}

	// Map task type
	var taskTypeEnum pb.TaskType
	switch strings.ToLower(taskType) {
	case "monitoring":
		taskTypeEnum = pb.TaskType_TASK_TYPE_MONITORING
	case "ai":
		taskTypeEnum = pb.TaskType_TASK_TYPE_AI
	case "communication":
		taskTypeEnum = pb.TaskType_TASK_TYPE_COMMUNICATION
	case "custom":
		taskTypeEnum = pb.TaskType_TASK_TYPE_CUSTOM
	default:
		fmt.Printf("Invalid task type: %s\n", taskType)
		return
	}

	// Map priority
	var priorityEnum pb.TaskPriority
	switch strings.ToLower(priority) {
	case "low":
		priorityEnum = pb.TaskPriority_TASK_PRIORITY_LOW
	case "medium":
		priorityEnum = pb.TaskPriority_TASK_PRIORITY_MEDIUM
	case "high":
		priorityEnum = pb.TaskPriority_TASK_PRIORITY_HIGH
	case "critical":
		priorityEnum = pb.TaskPriority_TASK_PRIORITY_CRITICAL
	default:
		fmt.Printf("Invalid priority: %s\n", priority)
		return
	}

	// Create task
	resp, err := taskClient.CreateTask(context.Background(), &pb.CreateTaskRequest{
		Type:     taskTypeEnum,
		Priority: priorityEnum,
		Payload:  payload,
	})

	if err != nil {
		fmt.Printf("Error creating task: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Failed to create task: %s\n", resp.Message)
		return
	}

	fmt.Printf("Task created successfully: %s\n", resp.Task.Id)
	fmt.Println()
}

// printHelp prints help information
func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  status    - Show cluster status")
	fmt.Println("  tasks     - Show all tasks")
	fmt.Println("  nodes     - Show all nodes")
	fmt.Println("  create-task - Create a new task")
	fmt.Println("  help      - Show this help")
	fmt.Println("  exit/quit - Exit the CLI")
	fmt.Println()
}
