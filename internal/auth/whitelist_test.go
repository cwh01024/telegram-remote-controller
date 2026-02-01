package auth

import (
	"testing"
)

func TestNewWhitelist(t *testing.T) {
	users := []int64{123, 456, 789}
	w := NewWhitelist(users)

	if w == nil {
		t.Fatal("NewWhitelist returned nil")
	}

	if len(w.GetAllowedUsers()) != 3 {
		t.Errorf("Expected 3 users, got %d", len(w.GetAllowedUsers()))
	}
}

func TestIsAuthorized(t *testing.T) {
	users := []int64{123, 456}
	w := NewWhitelist(users)

	tests := []struct {
		userID   int64
		expected bool
	}{
		{123, true},
		{456, true},
		{789, false},
		{0, false},
		{-1, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := w.IsAuthorized(tt.userID)
			if result != tt.expected {
				t.Errorf("IsAuthorized(%d) = %v, want %v", tt.userID, result, tt.expected)
			}
		})
	}
}

func TestAddRemoveUser(t *testing.T) {
	w := NewWhitelist([]int64{123})

	// Initially 123 is authorized
	if !w.IsAuthorized(123) {
		t.Error("User 123 should be authorized")
	}

	// Add new user
	w.AddUser(456)
	if !w.IsAuthorized(456) {
		t.Error("User 456 should be authorized after adding")
	}

	// Remove user
	w.RemoveUser(123)
	if w.IsAuthorized(123) {
		t.Error("User 123 should not be authorized after removal")
	}
}

func TestEmptyWhitelist(t *testing.T) {
	w := NewWhitelist([]int64{})

	if w.IsAuthorized(123) {
		t.Error("Empty whitelist should not authorize anyone")
	}
}
