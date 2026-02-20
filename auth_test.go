package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAuthService(t *testing.T) {
	auth, err := NewAuthService("allowed_users.cfg.example", "allowed_groups.cfg.example")
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}
	if auth == nil {
		t.Fatal("Expected auth service, got nil")
	}
	if auth.GetAllowedCount() != 2 {
		t.Errorf("Expected 2 allowed users, got %d", auth.GetAllowedCount())
	}
}

func TestNewAuthService_FileNotFound(t *testing.T) {
	_, err := NewAuthService("nonexistent_file.cfg", "allowed_groups.cfg.example")
	if err == nil {
		t.Error("Expected error for non-existent users file")
	}
}

func TestNewAuthService_GroupsFileNotFound(t *testing.T) {
	auth, err := NewAuthService("allowed_users.cfg.example", "nonexistent_groups.cfg")
	if err != nil {
		t.Errorf("Should not fail when groups file doesn't exist: %v", err)
	}
	if auth.GetAllowedGroupsCount() != 0 {
		t.Errorf("Expected 0 allowed groups, got %d", auth.GetAllowedGroupsCount())
	}
}

func TestAuthService_IsAllowed(t *testing.T) {
	auth, _ := NewAuthService("allowed_users.cfg.example", "allowed_groups.cfg.example")

	tests := []struct {
		name     string
		username string
		chatID   int64
		chatType string
		expected bool
	}{
		{"allowed user in private chat", "example_user1", 123, "private", true},
		{"allowed user with @ in private chat", "@example_user1", 123, "private", true},
		{"allowed user uppercase in private chat", "EXAMPLE_USER1", 123, "private", true},
		{"unknown user in private chat", "unknown_user", 123, "private", false},
		{"empty username in private chat", "", 123, "private", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.IsAllowed(tt.username, tt.chatID, tt.chatType, "")
			if result != tt.expected {
				t.Errorf("IsAllowed(%q, %d, %q) = %v, expected %v", tt.username, tt.chatID, tt.chatType, result, tt.expected)
			}
		})
	}
}

func TestAuthService_GroupAuthentication(t *testing.T) {
	tmpDir := t.TempDir()

	// Create users file
	usersFile := filepath.Join(tmpDir, "users.cfg")
	if err := os.WriteFile(usersFile, []byte("@user1\n"), 0644); err != nil {
		t.Fatalf("Failed to write users file: %v", err)
	}

	// Create groups file
	groupsFile := filepath.Join(tmpDir, "groups.cfg")
	groupsContent := "-1001234567890\nMy Test Group\n"
	if err := os.WriteFile(groupsFile, []byte(groupsContent), 0644); err != nil {
		t.Fatalf("Failed to write groups file: %v", err)
	}

	auth, err := NewAuthService(usersFile, groupsFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	tests := []struct {
		name      string
		username  string
		chatID    int64
		chatType  string
		chatTitle string
		expected  bool
	}{
		// Private chat tests
		{"allowed user private", "user1", 123, "private", "", true},
		{"unknown user private", "unknown", 123, "private", "", false},

		// Group chat tests - group ID match
		{"group ID allowed - unknown user", "unknown", -1001234567890, "supergroup", "Some Group", true},

		// Group chat tests - group title match
		{"group title allowed - unknown user", "unknown", -1009999999999, "group", "My Test Group", true},

		// Group chat tests - no match but allowed user
		{"unknown group - allowed user", "user1", -1009999999999, "supergroup", "Unknown Group", true},

		// Group chat tests - no match
		{"unknown group - unknown user", "unknown", -1009999999999, "supergroup", "Unknown Group", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.IsAllowed(tt.username, tt.chatID, tt.chatType, tt.chatTitle)
			if result != tt.expected {
				t.Errorf("IsAllowed(%q, %d, %q, %q) = %v, expected %v",
					tt.username, tt.chatID, tt.chatType, tt.chatTitle, result, tt.expected)
			}
		})
	}
}

func TestAuthService_IsGroupAllowed(t *testing.T) {
	tmpDir := t.TempDir()

	groupsFile := filepath.Join(tmpDir, "groups.cfg")
	groupsContent := "-1001234567890\nMy Test Group\nAnother Group\n"
	if err := os.WriteFile(groupsFile, []byte(groupsContent), 0644); err != nil {
		t.Fatalf("Failed to write groups file: %v", err)
	}

	auth, err := NewAuthService("allowed_users.cfg.example", groupsFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	tests := []struct {
		name      string
		chatID    int64
		chatTitle string
		expected  bool
	}{
		{"group ID match", -1001234567890, "Some Title", true},
		{"group title match 1", -1009999999999, "My Test Group", true},
		{"group title match 2", -1009999999999, "Another Group", true},
		{"no match - ID", -1009999999999, "", false},
		{"no match - title", -1009999999999, "Unknown Group", false},
		{"no match - empty title", -1009999999999, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.IsGroupAllowed(tt.chatID, tt.chatTitle)
			if result != tt.expected {
				t.Errorf("IsGroupAllowed(%d, %q) = %v, expected %v",
					tt.chatID, tt.chatTitle, result, tt.expected)
			}
		})
	}
}

func TestAuthService_GetAllowedUsers(t *testing.T) {
	auth, _ := NewAuthService("allowed_users.cfg.example", "allowed_groups.cfg.example")

	users := auth.GetAllowedUsers()
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestAuthService_GetAllowedGroups(t *testing.T) {
	tmpDir := t.TempDir()

	groupsFile := filepath.Join(tmpDir, "groups.cfg")
	groupsContent := "-1001234567890\nMy Group\n"
	if err := os.WriteFile(groupsFile, []byte(groupsContent), 0644); err != nil {
		t.Fatalf("Failed to write groups file: %v", err)
	}

	auth, err := NewAuthService("allowed_users.cfg.example", groupsFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	groups := auth.GetAllowedGroups()
	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}
}

func TestAuthService_Reload(t *testing.T) {
	tmpDir := t.TempDir()

	usersFile := filepath.Join(tmpDir, "test_users.cfg")
	groupsFile := filepath.Join(tmpDir, "test_groups.cfg")

	// Create initial files
	if err := os.WriteFile(usersFile, []byte("@user1\n@user2\n"), 0644); err != nil {
		t.Fatalf("Failed to write users file: %v", err)
	}
	if err := os.WriteFile(groupsFile, []byte("-1001111111111\n"), 0644); err != nil {
		t.Fatalf("Failed to write groups file: %v", err)
	}

	auth, err := NewAuthService(usersFile, groupsFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	// Verify initial state
	if auth.GetAllowedCount() != 2 {
		t.Errorf("Expected 2 users initially, got %d", auth.GetAllowedCount())
	}
	if auth.GetAllowedGroupsCount() != 1 {
		t.Errorf("Expected 1 group initially, got %d", auth.GetAllowedGroupsCount())
	}

	// Update files
	if err := os.WriteFile(usersFile, []byte("@user1\n@user3\n"), 0644); err != nil {
		t.Fatalf("Failed to update users file: %v", err)
	}
	if err := os.WriteFile(groupsFile, []byte("-1002222222222\n-1003333333333\n"), 0644); err != nil {
		t.Fatalf("Failed to update groups file: %v", err)
	}

	// Reload
	if err := auth.Reload(); err != nil {
		t.Fatalf("Failed to reload: %v", err)
	}

	// Verify reloaded state
	if auth.GetAllowedCount() != 2 {
		t.Errorf("Expected 2 users after reload, got %d", auth.GetAllowedCount())
	}
	if auth.GetAllowedGroupsCount() != 2 {
		t.Errorf("Expected 2 groups after reload, got %d", auth.GetAllowedGroupsCount())
	}
	if !auth.IsAllowed("user1", 123, "private", "") {
		t.Error("Expected user1 to still be allowed")
	}
	if auth.IsAllowed("user2", 123, "private", "") {
		t.Error("Expected user2 to no longer be allowed")
	}
	if !auth.IsAllowed("user3", 123, "private", "") {
		t.Error("Expected user3 to be allowed after reload")
	}
	if !auth.IsGroupAllowed(-1002222222222, "") {
		t.Error("Expected group -1002222222222 to be allowed after reload")
	}
}

func TestAuthService_EmptyLinesAndComments(t *testing.T) {
	tmpDir := t.TempDir()

	usersFile := filepath.Join(tmpDir, "test_users.cfg")
	groupsFile := filepath.Join(tmpDir, "test_groups.cfg")

	usersContent := "@user1\n\n# comment\n@user2\n   \n@user3\n"
	groupsContent := "-1001111111111\n\n# comment\n-1002222222222\n   \nMy Group\n"

	if err := os.WriteFile(usersFile, []byte(usersContent), 0644); err != nil {
		t.Fatalf("Failed to write users file: %v", err)
	}
	if err := os.WriteFile(groupsFile, []byte(groupsContent), 0644); err != nil {
		t.Fatalf("Failed to write groups file: %v", err)
	}

	auth, err := NewAuthService(usersFile, groupsFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	if auth.GetAllowedCount() != 3 {
		t.Errorf("Expected 3 users (ignoring empty lines and comments), got %d", auth.GetAllowedCount())
	}
	if auth.GetAllowedGroupsCount() != 3 {
		t.Errorf("Expected 3 groups (ignoring empty lines and comments), got %d", auth.GetAllowedGroupsCount())
	}
}

func TestAuthService_TrimsAtSymbol(t *testing.T) {
	tmpDir := t.TempDir()

	usersFile := filepath.Join(tmpDir, "test_users.cfg")
	groupsFile := filepath.Join(tmpDir, "test_groups.cfg")

	if err := os.WriteFile(usersFile, []byte("@testuser\n"), 0644); err != nil {
		t.Fatalf("Failed to write users file: %v", err)
	}
	if err := os.WriteFile(groupsFile, []byte("-1001234567890\n"), 0644); err != nil {
		t.Fatalf("Failed to write groups file: %v", err)
	}

	auth, err := NewAuthService(usersFile, groupsFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	if !auth.IsAllowed("testuser", 123, "private", "") {
		t.Error("Expected 'testuser' to be allowed")
	}
	if !auth.IsAllowed("@testuser", 123, "private", "") {
		t.Error("Expected '@testuser' to be allowed")
	}
}

func TestAuthService_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()

	usersFile := filepath.Join(tmpDir, "test_users.cfg")
	groupsFile := filepath.Join(tmpDir, "test_groups.cfg")

	if err := os.WriteFile(usersFile, []byte("@TestUser\n"), 0644); err != nil {
		t.Fatalf("Failed to write users file: %v", err)
	}
	if err := os.WriteFile(groupsFile, []byte("-1001234567890\n"), 0644); err != nil {
		t.Fatalf("Failed to write groups file: %v", err)
	}

	auth, err := NewAuthService(usersFile, groupsFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	tests := []string{"TestUser", "testuser", "TESTUSER", "@TestUser", "@testuser", "@TESTUSER"}
	for _, username := range tests {
		if !auth.IsAllowed(username, 123, "private", "") {
			t.Errorf("Expected %q to be allowed (case insensitive)", username)
		}
	}
}
