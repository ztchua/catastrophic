package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAuthService(t *testing.T) {
	auth, err := NewAuthService("allowed_users.cfg.example")
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
	_, err := NewAuthService("nonexistent_file.cfg")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestAuthService_IsAllowed(t *testing.T) {
	auth, _ := NewAuthService("allowed_users.cfg.example")

	tests := []struct {
		username string
		expected bool
	}{
		{"example_user1", true},
		{"@example_user1", true},
		{"EXAMPLE_USER1", true},
		{"@EXAMPLE_USER1", true},
		{"example_user2", true},
		{"@example_user2", true},
		{"unknown_user", false},
		{"@unknown_user", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			result := auth.IsAllowed(tt.username)
			if result != tt.expected {
				t.Errorf("IsAllowed(%q) = %v, expected %v", tt.username, result, tt.expected)
			}
		})
	}
}

func TestAuthService_GetAllowedUsers(t *testing.T) {
	auth, _ := NewAuthService("allowed_users.cfg.example")

	users := auth.GetAllowedUsers()
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestAuthService_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_users.cfg")

	initialContent := "@user1\n@user2\n"
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	auth, err := NewAuthService(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	if auth.GetAllowedCount() != 2 {
		t.Errorf("Expected 2 users initially, got %d", auth.GetAllowedCount())
	}

	if !auth.IsAllowed("user1") {
		t.Error("Expected user1 to be allowed")
	}
	if !auth.IsAllowed("user2") {
		t.Error("Expected user2 to be allowed")
	}

	updatedContent := "@user1\n@user3\n"
	if err := os.WriteFile(tmpFile, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update temp file: %v", err)
	}

	if err := auth.Reload(); err != nil {
		t.Fatalf("Failed to reload: %v", err)
	}

	if auth.GetAllowedCount() != 2 {
		t.Errorf("Expected 2 users after reload, got %d", auth.GetAllowedCount())
	}

	if !auth.IsAllowed("user1") {
		t.Error("Expected user1 to still be allowed")
	}
	if auth.IsAllowed("user2") {
		t.Error("Expected user2 to no longer be allowed")
	}
	if !auth.IsAllowed("user3") {
		t.Error("Expected user3 to be allowed after reload")
	}
}

func TestAuthService_EmptyLinesAndComments(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_users.cfg")

	content := "@user1\n\n# comment\n@user2\n   \n@user3\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	auth, err := NewAuthService(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	if auth.GetAllowedCount() != 3 {
		t.Errorf("Expected 3 users (ignoring empty lines and comments), got %d", auth.GetAllowedCount())
	}

	if !auth.IsAllowed("user1") {
		t.Error("Expected user1 to be allowed")
	}
	if !auth.IsAllowed("user2") {
		t.Error("Expected user2 to be allowed")
	}
	if !auth.IsAllowed("user3") {
		t.Error("Expected user3 to be allowed")
	}
}

func TestAuthService_TrimsAtSymbol(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_users.cfg")

	content := "@testuser\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	auth, err := NewAuthService(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	if !auth.IsAllowed("testuser") {
		t.Error("Expected 'testuser' to be allowed")
	}
	if !auth.IsAllowed("@testuser") {
		t.Error("Expected '@testuser' to be allowed")
	}
}

func TestAuthService_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_users.cfg")

	content := "@TestUser\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	auth, err := NewAuthService(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	tests := []string{"TestUser", "testuser", "TESTUSER", "@TestUser", "@testuser", "@TESTUSER"}
	for _, username := range tests {
		if !auth.IsAllowed(username) {
			t.Errorf("Expected %q to be allowed (case insensitive)", username)
		}
	}
}
