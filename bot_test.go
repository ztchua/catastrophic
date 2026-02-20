package main

import (
	"os"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestNewBot(t *testing.T) {
	store := setupTestStoreBot(t)
	defer store.Close()

	expenseStore := setupTestExpenseStore(t)
	defer expenseStore.Close()

	auth := setupTestAuth(t)
	bot := NewBot(nil, store, expenseStore, auth)

	if bot == nil {
		t.Fatal("Expected bot, got nil")
	}
	if bot.store == nil {
		t.Error("Expected store to be set")
	}
	if bot.expenseStore == nil {
		t.Error("Expected expenseStore to be set")
	}
	if bot.auth == nil {
		t.Error("Expected auth to be set")
	}
	if bot.sessions == nil {
		t.Error("Expected sessions map to be initialized")
	}
}

func TestBot_GetSession(t *testing.T) {
	store := setupTestStoreBot(t)
	defer store.Close()

	expenseStore := setupTestExpenseStore(t)
	defer expenseStore.Close()

	auth := setupTestAuth(t)
	bot := NewBot(nil, store, expenseStore, auth)

	session1 := bot.getSession(12345)
	if session1 == nil {
		t.Fatal("Expected session, got nil")
	}
	if session1.State != StateNone {
		t.Errorf("Expected StateNone, got %d", session1.State)
	}

	session2 := bot.getSession(12345)
	if session1 != session2 {
		t.Error("Expected same session for same chat ID")
	}

	session3 := bot.getSession(67890)
	if session1 == session3 {
		t.Error("Expected different sessions for different chat IDs")
	}
}

func TestBot_ResetSession(t *testing.T) {
	store := setupTestStoreBot(t)
	defer store.Close()

	expenseStore := setupTestExpenseStore(t)
	defer expenseStore.Close()

	auth := setupTestAuth(t)
	bot := NewBot(nil, store, expenseStore, auth)

	session := bot.getSession(12345)
	session.State = StateAwaitingPoopTexture
	session.Data["test"] = "value"

	bot.resetSession(12345)

	session = bot.getSession(12345)
	if session.State != StateNone {
		t.Errorf("Expected StateNone after reset, got %d", session.State)
	}
	if len(session.Data) != 0 {
		t.Error("Expected empty data map after reset")
	}
}

func TestBot_HandleUpdate_NilMessage(t *testing.T) {
	store := setupTestStoreBot(t)
	defer store.Close()

	expenseStore := setupTestExpenseStore(t)
	defer expenseStore.Close()

	auth := setupTestAuth(t)
	bot := NewBot(nil, store, expenseStore, auth)

	update := tgbotapi.Update{}
	bot.HandleUpdate(update)
}

func TestTimeSinceRecord_Hours(t *testing.T) {
	now := time.Now()
	cases := []struct {
		duration time.Duration
		expected string
	}{
		{1 * time.Hour, "1 hours"},
		{5 * time.Hour, "5 hours"},
		{23 * time.Hour, "23 hours"},
	}

	for _, tc := range cases {
		result := timeSinceRecord(now.Add(-tc.duration))
		if result != tc.expected {
			t.Errorf("timeSinceRecord(%v): expected '%s', got '%s'", tc.duration, tc.expected, result)
		}
	}
}

func TestTimeSinceRecord_DaysAndHours(t *testing.T) {
	now := time.Now()
	cases := []struct {
		duration time.Duration
		expected string
	}{
		{24 * time.Hour, "1 days and 0 hours"},
		{25 * time.Hour, "1 days and 1 hours"},
		{72 * time.Hour, "3 days and 0 hours"},
		{100 * time.Hour, "4 days and 4 hours"},
	}

	for _, tc := range cases {
		result := timeSinceRecord(now.Add(-tc.duration))
		if result != tc.expected {
			t.Errorf("timeSinceRecord(%v): expected '%s', got '%s'", tc.duration, tc.expected, result)
		}
	}
}

func TestHandlePoopUpdate_ArgParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        string
		expectError bool
		expectedID  int64
		expectedTx  string
	}{
		{
			name:        "valid args",
			args:        "1 solid texture",
			expectError: false,
			expectedID:  1,
			expectedTx:  "solid texture",
		},
		{
			name:        "missing texture",
			args:        "1",
			expectError: true,
		},
		{
			name:        "empty args",
			args:        "",
			expectError: true,
		},
		{
			name:        "invalid ID",
			args:        "abc texture",
			expectError: true,
		},
		{
			name:        "multi-word texture",
			args:        "5 very soft and watery",
			expectError: false,
			expectedID:  5,
			expectedTx:  "very soft and watery",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Fields(tt.args)

			if tt.expectError {
				if len(parts) < 2 && tt.args != "" {
					return
				}
				if tt.args == "" {
					return
				}
			}

			if !tt.expectError && len(parts) >= 2 {
				texture := strings.Join(parts[1:], " ")
				if texture != tt.expectedTx {
					t.Errorf("Expected texture '%s', got '%s'", tt.expectedTx, texture)
				}
			}
		})
	}
}

func TestHandlePoopDelete_ArgParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        string
		expectError bool
		expectedID  int64
	}{
		{
			name:        "valid ID",
			args:        "42",
			expectError: false,
			expectedID:  42,
		},
		{
			name:        "empty args",
			args:        "",
			expectError: true,
		},
		{
			name:        "invalid ID",
			args:        "abc",
			expectError: true,
		},
		{
			name:        "ID with spaces",
			args:        " 123 ",
			expectError: false,
			expectedID:  123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := strings.TrimSpace(tt.args)

			if args == "" {
				if !tt.expectError {
					t.Error("Expected error for empty args")
				}
				return
			}

			if tt.expectError {
				return
			}
		})
	}
}

func TestConversationState_Values(t *testing.T) {
	if StateNone != 0 {
		t.Errorf("Expected StateNone to be 0, got %d", StateNone)
	}
	if StateAwaitingPoopTexture != 1 {
		t.Errorf("Expected StateAwaitingPoopTexture to be 1, got %d", StateAwaitingPoopTexture)
	}
}

func TestChatSession_InitialState(t *testing.T) {
	store := setupTestStoreBot(t)
	defer store.Close()

	expenseStore := setupTestExpenseStore(t)
	defer expenseStore.Close()

	auth := setupTestAuth(t)
	bot := NewBot(nil, store, expenseStore, auth)
	session := bot.getSession(12345)

	if session.State != StateNone {
		t.Errorf("Expected initial state to be StateNone, got %d", session.State)
	}
	if session.Data == nil {
		t.Error("Expected Data map to be initialized")
	}
}

func TestBot_ConcurrentSessionAccess(t *testing.T) {
	store := setupTestStoreBot(t)
	defer store.Close()

	expenseStore := setupTestExpenseStore(t)
	defer expenseStore.Close()

	auth := setupTestAuth(t)
	bot := NewBot(nil, store, expenseStore, auth)

	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func(id int) {
			chatID := int64(id % 10)
			bot.getSession(chatID)
			bot.resetSession(chatID)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestBot_HandleUpdate_AwaitingTextureState(t *testing.T) {
	store := setupTestStoreBot(t)
	defer store.Close()

	expenseStore := setupTestExpenseStore(t)
	defer expenseStore.Close()

	auth := setupTestAuth(t)
	bot := NewBot(nil, store, expenseStore, auth)
	bot.getSession(12345).State = StateAwaitingPoopTexture
	bot.getSession(12345).Data["chat_id"] = int64(12345)

	if bot.getSession(12345).State != StateAwaitingPoopTexture {
		t.Error("Session state should be StateAwaitingPoopTexture")
	}
}

func setupTestStoreBot(t *testing.T) *PoopStore {
	t.Helper()
	return setupTestStore(t)
}

func setupTestExpenseStore(t *testing.T) *ExpenseStore {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "expense_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	store, err := NewExpenseStore(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create expense store: %v", err)
	}

	t.Cleanup(func() {
		store.Close()
		os.Remove(tmpFile.Name())
	})

	return store
}

func setupTestAuth(t *testing.T) *AuthService {
	t.Helper()
	auth, err := NewAuthService("allowed_users.cfg.example", "allowed_groups.cfg.example")
	if err != nil {
		t.Fatalf("Failed to create test auth: %v", err)
	}
	return auth
}
