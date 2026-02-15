package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type AuthService struct {
	allowedUsers map[string]bool
	filePath     string
	mu           sync.RWMutex
}

func NewAuthService(filePath string) (*AuthService, error) {
	auth := &AuthService{
		allowedUsers: make(map[string]bool),
		filePath:     filePath,
	}

	if err := auth.loadAllowedUsers(); err != nil {
		return nil, err
	}

	return auth, nil
}

func (a *AuthService) loadAllowedUsers() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.Open(a.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("allowed users config file not found: %s", a.filePath)
		}
		return fmt.Errorf("failed to open allowed users file: %w", err)
	}
	defer file.Close()

	a.allowedUsers = make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		username := strings.TrimPrefix(line, "@")
		username = strings.ToLower(username)
		if username != "" {
			a.allowedUsers[username] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read allowed users file: %w", err)
	}

	return nil
}

func (a *AuthService) Reload() error {
	return a.loadAllowedUsers()
}

func (a *AuthService) IsAllowed(username string) bool {
	if username == "" {
		return false
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	normalizedUsername := strings.ToLower(strings.TrimPrefix(username, "@"))
	return a.allowedUsers[normalizedUsername]
}

func (a *AuthService) GetAllowedUsers() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	users := make([]string, 0, len(a.allowedUsers))
	for user := range a.allowedUsers {
		users = append(users, "@"+user)
	}
	return users
}

func (a *AuthService) GetAllowedCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.allowedUsers)
}
