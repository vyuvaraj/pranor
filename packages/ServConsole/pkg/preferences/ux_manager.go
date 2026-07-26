package preferences

import (
	"encoding/json"
	"net/http"
	"sync"
)

// UserUXPreferences stores theme, layout, pinned widgets, and custom keyboard shortcuts.
type UserUXPreferences struct {
	UserID            string            `json:"user_id"`
	Theme             string            `json:"theme"` // "dark", "light", "glassmorphism"
	PinnedWidgets     []string          `json:"pinned_widgets"`
	KeyboardShortcuts map[string]string `json:"keyboard_shortcuts"` // action -> shortcut (e.g. "open_search" -> "Cmd+K")
}

// UXPreferencesManager handles user UX customization settings for ServConsole.
type UXPreferencesManager struct {
	mu          sync.RWMutex
	preferences map[string]*UserUXPreferences // userID -> preferences
}

// NewUXPreferencesManager creates a UXPreferencesManager instance.
func NewUXPreferencesManager() *UXPreferencesManager {
	return &UXPreferencesManager{
		preferences: make(map[string]*UserUXPreferences),
	}
}

// SavePreferences updates user preferences.
func (upm *UXPreferencesManager) SavePreferences(pref UserUXPreferences) *UserUXPreferences {
	if pref.UserID == "" {
		pref.UserID = "default-user"
	}
	if pref.Theme == "" {
		pref.Theme = "glassmorphism"
	}
	if pref.KeyboardShortcuts == nil {
		pref.KeyboardShortcuts = map[string]string{
			"search": "Cmd+K",
			"home":   "G+H",
		}
	}

	upm.mu.Lock()
	defer upm.mu.Unlock()

	upm.preferences[pref.UserID] = &pref
	return &pref
}

// GetPreferences retrieves preferences for a user ID.
func (upm *UXPreferencesManager) GetPreferences(userID string) (*UserUXPreferences, bool) {
	if userID == "" {
		userID = "default-user"
	}
	upm.mu.RLock()
	defer upm.mu.RUnlock()
	pref, ok := upm.preferences[userID]
	return pref, ok
}

// HTTPHandler exposes `/api/v1/console/preferences` for ServConsole settings UI.
func (upm *UXPreferencesManager) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost {
			var pref UserUXPreferences
			if err := json.NewDecoder(r.Body).Decode(&pref); err != nil {
				http.Error(w, "invalid JSON payload", http.StatusBadRequest)
				return
			}
			saved := upm.SavePreferences(pref)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(saved)
			return
		}

		userID := r.URL.Query().Get("user_id")
		pref, found := upm.GetPreferences(userID)
		if !found {
			pref = upm.SavePreferences(UserUXPreferences{UserID: userID})
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(pref)
	})
}
