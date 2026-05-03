package common

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

// SecurityManager manages security-related operations
type SecurityManager struct {
	apiKeys        map[string]string
	allowedCommands []string
	tlsConfig      *tls.Config
}

// NewSecurityManager creates a new SecurityManager instance
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		apiKeys:        make(map[string]string),
		allowedCommands: []string{},
		tlsConfig:      &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}

// LoadAPIKeys loads API keys from environment variables or a .env file
func (sm *SecurityManager) LoadAPIKeys() error {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load API keys from environment variables
	apiKeyNames := []string{
		"OPENAI_API_KEY",
		"ANTHROPIC_API_KEY",
		"GOOGLE_API_KEY",
		"TELEGRAM_BOT_TOKEN",
		"DISCORD_BOT_TOKEN",
	}

	for _, keyName := range apiKeyNames {
		if key := os.Getenv(keyName); key != "" {
			sm.apiKeys[keyName] = key
		}
	}

	return nil
}

// GetAPIKey returns an API key for the specified service
func (sm *SecurityManager) GetAPIKey(service string) (string, error) {
	key, exists := sm.apiKeys[service]
	if !exists {
		return "", fmt.Errorf("API key not found for service: %s", service)
	}
	return key, nil
}

// ValidateCommand validates if a command is allowed
func (sm *SecurityManager) ValidateCommand(command string) bool {
	// If no allowed commands are specified, allow all commands
	if len(sm.allowedCommands) == 0 {
		return true
	}

	// Check if the command is in the allowed list
	for _, allowedCommand := range sm.allowedCommands {
		if strings.HasPrefix(command, allowedCommand) {
			return true
		}
	}

	return false
}

// SetAllowedCommands sets the list of allowed commands
func (sm *SecurityManager) SetAllowedCommands(commands []string) {
	sm.allowedCommands = commands
}

// SanitizeInput sanitizes user input to prevent injection attacks
func (sm *SecurityManager) SanitizeInput(input string) string {
	// Remove potentially harmful characters
	sanitized := input
	sanitized = regexp.MustCompile(`[;<>|&]`).ReplaceAllString(sanitized, "")
	return sanitized
}

// SecurePath ensures a path is within the allowed directory
func (sm *SecurityManager) SecurePath(baseDir, requestedPath string) (string, error) {
	// Get absolute paths
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}

	requestedAbs, err := filepath.Abs(filepath.Join(baseDir, requestedPath))
	if err != nil {
		return "", err
	}

	// Check if the requested path is within the base directory
	if !strings.HasPrefix(requestedAbs, baseAbs) {
		return "", fmt.Errorf("path traversal detected: %s", requestedPath)
	}

	return requestedAbs, nil
}

// CreateHTTPClient creates a secure HTTP client
func (sm *SecurityManager) CreateHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: sm.tlsConfig,
		},
	}
}

// CheckFilePermissions checks if a file has secure permissions
func (sm *SecurityManager) CheckFilePermissions(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Check if the file is readable only by the owner
	mode := info.Mode()
	if mode&0077 != 0 {
		return fmt.Errorf("file %s has insecure permissions: %s", filePath, mode)
	}

	return nil
}
