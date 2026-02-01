package auth

import (
	"log"
	"sync"
)

// Authenticator verifies if a user is allowed to use the bot
type Authenticator interface {
	IsAuthorized(userID int64) bool
	GetAllowedUsers() []int64
}

// Whitelist implements Authenticator with a list of allowed user IDs
type Whitelist struct {
	mu           sync.RWMutex
	allowedUsers map[int64]bool
}

// NewWhitelist creates a new Whitelist authenticator
func NewWhitelist(userIDs []int64) *Whitelist {
	w := &Whitelist{
		allowedUsers: make(map[int64]bool),
	}
	for _, id := range userIDs {
		w.allowedUsers[id] = true
	}
	return w
}

// IsAuthorized checks if the user ID is in the whitelist
func (w *Whitelist) IsAuthorized(userID int64) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	allowed := w.allowedUsers[userID]
	if !allowed {
		log.Printf("Unauthorized access attempt from user ID: %d", userID)
	}
	return allowed
}

// GetAllowedUsers returns a copy of all allowed user IDs
func (w *Whitelist) GetAllowedUsers() []int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	users := make([]int64, 0, len(w.allowedUsers))
	for id := range w.allowedUsers {
		users = append(users, id)
	}
	return users
}

// AddUser adds a user to the whitelist
func (w *Whitelist) AddUser(userID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.allowedUsers[userID] = true
	log.Printf("Added user %d to whitelist", userID)
}

// RemoveUser removes a user from the whitelist
func (w *Whitelist) RemoveUser(userID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.allowedUsers, userID)
	log.Printf("Removed user %d from whitelist", userID)
}
