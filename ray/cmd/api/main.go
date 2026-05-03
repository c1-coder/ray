package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"ray/src/common"
)

var (
	port = flag.Int("port", 8080, "API server port")
	masterAddr = flag.String("master", "localhost:50051", "Master node address")
)

type API struct {
	serviceRegistry *common.ServiceRegistry
	taskQueue       *common.TaskQueue
}

func main() {
	flag.Parse()

	// Create service registry
	serviceRegistry := common.NewServiceRegistry(30 * time.Second)

	// Create task queue
	taskQueue := common.NewTaskQueue()

	// Create API instance
	api := &API{
		serviceRegistry: serviceRegistry,
		taskQueue:       taskQueue,
	}

	// Set up Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	apiGroup := router.Group("/api")
	{
		// Dashboard
		apiGroup.GET("/dashboard", api.getDashboard)

		// Tasks
		apiGroup.GET("/tasks", api.getTasks)
		apiGroup.POST("/tasks", api.createTask)
		apiGroup.PUT("/tasks/:id", api.updateTask)
		apiGroup.DELETE("/tasks/:id", api.deleteTask)

		// Nodes
		apiGroup.GET("/nodes", api.getNodes)
		apiGroup.POST("/nodes", api.createNode)
		apiGroup.PUT("/nodes/:id", api.updateNode)
		apiGroup.DELETE("/nodes/:id", api.deleteNode)
		apiGroup.POST("/nodes/:id/restart", api.restartNode)

		// Security
		apiGroup.GET("/security", api.getSecurity)
		apiGroup.POST("/security/scan", api.scanSecurity)
		apiGroup.POST("/security/api-keys", api.generateAPIKey)

		// Settings
		apiGroup.GET("/settings", api.getSettings)
		apiGroup.PUT("/settings", api.updateSettings)
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("API server starting on port %d", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// getDashboard returns dashboard data
func (api *API) getDashboard(c *gin.Context) {
	// Mock dashboard data
	data := map[string]interface{}{
		"stats": map[string]int{
			"totalTasks":      150,
			"completedTasks":  120,
			"pendingTasks":    30,
			"totalNodes":      4,
			"onlineNodes":     3,
			"offlineNodes":    1,
			"securityScore":   85,
			"totalVulnerabilities": 4,
			"fixedVulnerabilities": 3,
		},
		"recentTasks": []map[string]interface{}{
			{
				"id":        "1",
				"type":      "Monitoring",
				"status":    "completed",
				"priority":  "high",
				"assignedTo": "node-1",
				"createdAt": "2024-01-01 10:00:00",
			},
			{
				"id":        "2",
				"type":      "AI",
				"status":    "running",
				"priority":  "medium",
				"assignedTo": "node-2",
				"createdAt": "2024-01-01 11:00:00",
			},
			{
				"id":        "3",
				"type":      "Communication",
				"status":    "pending",
				"priority":  "low",
				"assignedTo": "",
				"createdAt": "2024-01-01 12:00:00",
			},
		},
		"nodeStatus": []map[string]interface{}{
			{
				"id":           "node-1",
				"type":         "worker",
				"address":      "192.168.1.101:50052",
				"status":       "online",
				"load":         45,
				"capabilities": []string{"task_execution", "monitoring"},
			},
			{
				"id":           "node-2",
				"type":         "worker",
				"address":      "192.168.1.102:50052",
				"status":       "online",
				"load":         65,
				"capabilities": []string{"task_execution", "ai_assistance"},
			},
			{
				"id":           "node-3",
				"type":         "worker",
				"address":      "192.168.1.103:50052",
				"status":       "online",
				"load":         25,
				"capabilities": []string{"task_execution", "communication"},
			},
			{
				"id":           "node-4",
				"type":         "worker",
				"address":      "192.168.1.104:50052",
				"status":       "offline",
				"load":         0,
				"capabilities": []string{"task_execution"},
			},
		},
	}

	c.JSON(http.StatusOK, data)
}

// getTasks returns all tasks
func (api *API) getTasks(c *gin.Context) {
	// Mock tasks data
	tasks := []map[string]interface{}{
		{
			"id":        "1",
			"type":      "Monitoring",
			"status":    "completed",
			"priority":  "high",
			"assignedTo": "node-1",
			"payload":   `{"target": "server-1", "metrics": ["cpu", "memory", "disk"]}`,
			"createdAt": "2024-01-01 10:00:00",
			"updatedAt": "2024-01-01 10:30:00",
		},
		{
			"id":        "2",
			"type":      "AI",
			"status":    "running",
			"priority":  "medium",
			"assignedTo": "node-2",
			"payload":   `{"prompt": "Generate a summary of the latest news", "model": "gpt-4"}`,
			"createdAt": "2024-01-01 11:00:00",
			"updatedAt": "2024-01-01 11:15:00",
		},
		{
			"id":        "3",
			"type":      "Communication",
			"status":    "pending",
			"priority":  "low",
			"assignedTo": "",
			"payload":   `{"recipient": "user@example.com", "message": "Hello from Ray Platform"}`,
			"createdAt": "2024-01-01 12:00:00",
			"updatedAt": "2024-01-01 12:00:00",
		},
		{
			"id":        "4",
			"type":      "Monitoring",
			"status":    "completed",
			"priority":  "medium",
			"assignedTo": "node-3",
			"payload":   `{"target": "server-2", "metrics": ["cpu", "memory"]}`,
			"createdAt": "2024-01-01 09:00:00",
			"updatedAt": "2024-01-01 09:30:00",
		},
		{
			"id":        "5",
			"type":      "AI",
			"status":    "pending",
			"priority":  "high",
			"assignedTo": "",
			"payload":   `{"prompt": "Analyze the sales data", "model": "gpt-4"}`,
			"createdAt": "2024-01-01 13:00:00",
			"updatedAt": "2024-01-01 13:00:00",
		},
	}

	c.JSON(http.StatusOK, tasks)
}

// createTask creates a new task
func (api *API) createTask(c *gin.Context) {
	var taskData map[string]interface{}
	if err := c.ShouldBindJSON(&taskData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create new task
	newTask := map[string]interface{}{
		"id":        fmt.Sprintf("%d", time.Now().UnixNano()),
		"type":      taskData["type"],
		"status":    "pending",
		"priority":  taskData["priority"],
		"assignedTo": taskData["assignedTo"],
		"payload":   taskData["payload"],
		"createdAt": time.Now().Format("2006-01-02 15:04:05"),
		"updatedAt": time.Now().Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusCreated, newTask)
}

// updateTask updates an existing task
func (api *API) updateTask(c *gin.Context) {
	id := c.Param("id")
	var taskData map[string]interface{}
	if err := c.ShouldBindJSON(&taskData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update task
	updatedTask := map[string]interface{}{
		"id":        id,
		"type":      taskData["type"],
		"status":    taskData["status"],
		"priority":  taskData["priority"],
		"assignedTo": taskData["assignedTo"],
		"payload":   taskData["payload"],
		"createdAt": time.Now().Format("2006-01-02 15:04:05"),
		"updatedAt": time.Now().Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, updatedTask)
}

// deleteTask deletes a task
func (api *API) deleteTask(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Task %s deleted successfully", id)})
}

// getNodes returns all nodes
func (api *API) getNodes(c *gin.Context) {
	// Mock nodes data
	nodes := []map[string]interface{}{
		{
			"id":            "node-1",
			"type":          "worker",
			"address":       "192.168.1.101:50052",
			"status":        "online",
			"load":          45,
			"uptime":        "24h 30m",
			"capabilities":  []string{"task_execution", "monitoring"},
			"lastHeartbeat": "2024-01-01 10:30:00",
		},
		{
			"id":            "node-2",
			"type":          "worker",
			"address":       "192.168.1.102:50052",
			"status":        "online",
			"load":          65,
			"uptime":        "12h 15m",
			"capabilities":  []string{"task_execution", "ai_assistance"},
			"lastHeartbeat": "2024-01-01 10:29:00",
		},
		{
			"id":            "node-3",
			"type":          "worker",
			"address":       "192.168.1.103:50052",
			"status":        "online",
			"load":          25,
			"uptime":        "48h 05m",
			"capabilities":  []string{"task_execution", "communication"},
			"lastHeartbeat": "2024-01-01 10:30:00",
		},
		{
			"id":            "node-4",
			"type":          "worker",
			"address":       "192.168.1.104:50052",
			"status":        "offline",
			"load":          0,
			"uptime":        "0",
			"capabilities":  []string{"task_execution"},
			"lastHeartbeat": "2024-01-01 08:15:00",
		},
	}

	c.JSON(http.StatusOK, nodes)
}

// createNode creates a new node
func (api *API) createNode(c *gin.Context) {
	var nodeData map[string]interface{}
	if err := c.ShouldBindJSON(&nodeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create new node
	newNode := map[string]interface{}{
		"id":            fmt.Sprintf("node-%d", time.Now().UnixNano()),
		"type":          nodeData["type"],
		"address":       nodeData["address"],
		"status":        "connecting",
		"load":          0,
		"uptime":        "0",
		"capabilities":  nodeData["capabilities"],
		"lastHeartbeat": time.Now().Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusCreated, newNode)
}

// updateNode updates an existing node
func (api *API) updateNode(c *gin.Context) {
	id := c.Param("id")
	var nodeData map[string]interface{}
	if err := c.ShouldBindJSON(&nodeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update node
	updatedNode := map[string]interface{}{
		"id":            id,
		"type":          nodeData["type"],
		"address":       nodeData["address"],
		"status":        "online",
		"load":          0,
		"uptime":        "0",
		"capabilities":  nodeData["capabilities"],
		"lastHeartbeat": time.Now().Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, updatedNode)
}

// deleteNode deletes a node
func (api *API) deleteNode(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Node %s deleted successfully", id)})
}

// restartNode restarts a node
func (api *API) restartNode(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Node %s restarting...", id)})
}

// getSecurity returns security data
func (api *API) getSecurity(c *gin.Context) {
	// Mock security data
	data := map[string]interface{}{
		"securityScore": 85,
		"vulnerabilities": []map[string]interface{}{
			{
				"id":          "1",
				"tool":        "Nezha",
				"severity":    "medium",
				"description": "Potential XSS vulnerability in dashboard",
				"status":      "fixed",
				"fix":         "Updated to latest version",
			},
			{
				"id":          "2",
				"tool":        "Hermes Agent",
				"severity":    "low",
				"description": "Insecure password storage",
				"status":      "fixed",
				"fix":         "Implemented proper encryption",
			},
			{
				"id":          "3",
				"tool":        "OpenClaw",
				"severity":    "high",
				"description": "Remote code execution vulnerability",
				"status":      "pending",
				"fix":         "Update to version 1.2.0",
			},
			{
				"id":          "4",
				"tool":        "Ray Platform",
				"severity":    "low",
				"description": "Missing rate limiting on API endpoints",
				"status":      "fixed",
				"fix":         "Implemented rate limiting",
			},
		},
		"apiKeys": []map[string]interface{}{
			{
				"id":         "1",
				"name":       "Admin Key",
				"key":        "sk_admin_1234567890abcdef",
				"createdAt":  "2024-01-01 10:00:00",
				"lastUsed":   "2024-01-01 11:30:00",
				"permissions": []string{"read", "write", "admin"},
			},
			{
				"id":         "2",
				"name":       "Read-Only Key",
				"key":        "sk_read_1234567890abcdef",
				"createdAt":  "2024-01-01 09:00:00",
				"lastUsed":   "2024-01-01 10:15:00",
				"permissions": []string{"read"},
			},
		},
	}

	c.JSON(http.StatusOK, data)
}

// scanSecurity starts a security scan
func (api *API) scanSecurity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Security scan started", "status": "in_progress"})
}

// generateAPIKey generates a new API key
func (api *API) generateAPIKey(c *gin.Context) {
	var keyData map[string]interface{}
	if err := c.ShouldBindJSON(&keyData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate new API key
	newKey := map[string]interface{}{
		"id":         fmt.Sprintf("%d", time.Now().UnixNano()),
		"name":       keyData["name"],
		"key":        fmt.Sprintf("sk_%s_%s", keyData["name"], fmt.Sprintf("%x", time.Now().UnixNano())),
		"createdAt":  time.Now().Format("2006-01-02 15:04:05"),
		"lastUsed":   "",
		"permissions": keyData["permissions"],
	}

	c.JSON(http.StatusCreated, newKey)
}

// getSettings returns platform settings
func (api *API) getSettings(c *gin.Context) {
	// Mock settings data
	settings := map[string]interface{}{
		"masterAddress":      "localhost:50051",
		"heartbeatInterval":  30,
		"taskTimeout":        300,
		"maxRetries":         3,
		"enableTLS":          false,
		"enableAuthentication": true,
		"logLevel":           "info",
		"nodeCapabilities":   []string{"task_execution", "monitoring"},
		"toolIntegrations": map[string]interface{}{
			"nezha": map[string]interface{}{
				"path":   "/workspace/nezha",
				"status": "connected",
			},
			"hermesAgent": map[string]interface{}{
				"path":   "/workspace/hermes-agent",
				"status": "connected",
			},
			"openclaw": map[string]interface{}{
				"path":   "/workspace/openclaw",
				"status": "connected",
			},
		},
	}

	c.JSON(http.StatusOK, settings)
}

// updateSettings updates platform settings
func (api *API) updateSettings(c *gin.Context) {
	var settingsData map[string]interface{}
	if err := c.ShouldBindJSON(&settingsData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update settings
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully", "settings": settingsData})
}
