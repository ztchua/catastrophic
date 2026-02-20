package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type AuthService struct {
	allowedUsers   map[string]bool
	allowedGroups  map[string]bool
	usersFilePath  string
	groupsFilePath string
	mu             sync.RWMutex
}

func NewAuthService(usersFilePath string, groupsFilePath string) (*AuthService, error) {
	auth := &AuthService{
		allowedUsers:   make(map[string]bool),
		allowedGroups:  make(map[string]bool),
		usersFilePath:  usersFilePath,
		groupsFilePath: groupsFilePath,
	}

	if err := auth.loadAllowedUsers(); err != nil {
		return nil, err
	}

	// Load groups - if file doesn't exist, just use empty groups
	if err := auth.loadAllowedGroups(); err != nil {
		// Log but don't fail - groups are optional
		fmt.Printf("Warning: %v\n", err)
	}

	return auth, nil
}

func (a *AuthService) loadAllowedUsers() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.Open(a.usersFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("allowed users config file not found: %s", a.usersFilePath)
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

func (a *AuthService) loadAllowedGroups() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.Open(a.groupsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("allowed groups config file not found: %s", a.groupsFilePath)
		}
		return fmt.Errorf("failed to open allowed groups file: %w", err)
	}
	defer file.Close()

	a.allowedGroups = make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Store both the raw line and normalized versions
		// For group IDs (negative numbers), store as string
		// For group names, store as-is (case-sensitive)
		if line != "" {
			a.allowedGroups[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read allowed groups file: %w", err)
	}

	return nil
}

func (a *AuthService) Reload() error {
	if err := a.loadAllowedUsers(); err != nil {
		return err
	}
	// Try to reload groups, but don't fail if file doesn't exist
	if err := a.loadAllowedGroups(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}
	return nil
}

// IsAllowed checks if a user is allowed to use the bot.
// For private chats, it checks the username against allowedUsers.
// For group/supergroup chats, it first checks if the group is allowed,
// then falls back to checking the individual username.
func (a *AuthService) IsAllowed(username string, chatID int64, chatType string, chatTitle string) bool {
	// For private chats, only check username
	if chatType == "private" {
		return a.isUserAllowed(username)
	}

	// For groups/supergroups, check if group is allowed first
	if a.IsGroupAllowed(chatID, chatTitle) {
		return true
	}

	// Fall back to checking individual username
	return a.isUserAllowed(username)
}

// isUserAllowed checks if username is in the allowed users list
func (a *AuthService) isUserAllowed(username string) bool {
	if username == "" {
		return false
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	normalizedUsername := strings.ToLower(strings.TrimPrefix(username, "@"))
	return a.allowedUsers[normalizedUsername]
}

// IsGroupAllowed checks if a group (by ID or title) is in the allowed groups list
func (a *AuthService) IsGroupAllowed(chatID int64, chatTitle string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Check by chat ID (stored as string)
	chatIDStr := strconv.FormatInt(chatID, 10)
	if a.allowedGroups[chatIDStr] {
		return true
	}

	// Check by title (if provided)
	if chatTitle != "" && a.allowedGroups[chatTitle] {
		return true
	}

	return false
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

func (a *AuthService) GetAllowedGroupsCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.allowedGroups)
}

func (a *AuthService) GetAllowedGroups() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	groups := make([]string, 0, len(a.allowedGroups))
	for group := range a.allowedGroups {
		groups = append(groups, group)
	}
	return groups
}
