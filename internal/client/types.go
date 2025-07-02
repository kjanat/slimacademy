package client

import "time"

// Credentials holds the authentication information
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest matches the bash script's JSON payload structure
type LoginRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	DeviceID     string `json:"device_id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

// LoginResponse represents the authentication response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// TokenInfo stores token metadata for persistence
type TokenInfo struct {
	Token        string    `json:"token"`
	TokenType    string    `json:"token_type"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresIn    int       `json:"expires_in_seconds"`
	ExpiresAt    time.Time `json:"expires_at"`
	Username     string    `json:"username"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

// UserProfile represents the user profile response
type UserProfile struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	// Add other profile fields as needed
}

// LibraryResponse represents the library endpoint response
type LibraryResponse struct {
	Summaries []LibrarySummary `json:"summaries"`
	Total     int              `json:"total"`
}

// LibrarySummary represents a summary in the library
type LibrarySummary struct {
	ID             int      `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	LastOpenedAt   *string  `json:"lastOpenedAt"`
	ReadProgress   *int     `json:"readProgress"`
	ReadPercentage *float64 `json:"readPercentage"`
	// Add other fields as needed
}

// BookData holds all the data for a single book
type BookData struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Summary  any    `json:"summary"`
	Chapters any    `json:"chapters"`
	Content  any    `json:"content"`
	Notes    any    `json:"notes"`
}

// APIError represents an API error response
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Response   string `json:"response"`
}

func (e *APIError) Error() string {
	return e.Message
}
