package client

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCredentials_JSONMarshalUnmarshal(t *testing.T) {
	original := &Credentials{
		Username: "test@example.com",
		Password: "secretpassword123",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal credentials: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	if !strings.Contains(jsonStr, "test@example.com") {
		t.Error("JSON should contain username")
	}
	if !strings.Contains(jsonStr, "secretpassword123") {
		t.Error("JSON should contain password")
	}

	// Unmarshal from JSON
	var unmarshaled Credentials
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal credentials: %v", err)
	}

	// Compare fields
	if unmarshaled.Username != original.Username {
		t.Errorf("Expected username %q, got %q", original.Username, unmarshaled.Username)
	}
	if unmarshaled.Password != original.Password {
		t.Errorf("Expected password %q, got %q", original.Password, unmarshaled.Password)
	}
}

func TestCredentials_EmptyValues(t *testing.T) {
	creds := &Credentials{
		Username: "",
		Password: "",
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Failed to marshal empty credentials: %v", err)
	}

	var unmarshaled Credentials
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal empty credentials: %v", err)
	}

	if unmarshaled.Username != "" {
		t.Errorf("Expected empty username, got %q", unmarshaled.Username)
	}
	if unmarshaled.Password != "" {
		t.Errorf("Expected empty password, got %q", unmarshaled.Password)
	}
}

func TestLoginRequest_JSONMarshalUnmarshal(t *testing.T) {
	original := &LoginRequest{
		GrantType:    "external_password",
		ClientID:     "slim_api",
		ClientSecret: "",
		DeviceID:     "4b3c1096-114f-423b-822f-22785cc3f05e",
		Username:     "test@example.com",
		Password:     "mypassword123",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal login request: %v", err)
	}

	// Verify JSON structure
	jsonStr := string(data)
	expectedFields := []string{
		"grant_type", "client_id", "client_secret", "device_id", "username", "password",
	}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON should contain field %q", field)
		}
	}

	// Unmarshal from JSON
	var unmarshaled LoginRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal login request: %v", err)
	}

	// Compare all fields
	if unmarshaled.GrantType != original.GrantType {
		t.Errorf("Expected grant_type %q, got %q", original.GrantType, unmarshaled.GrantType)
	}
	if unmarshaled.ClientID != original.ClientID {
		t.Errorf("Expected client_id %q, got %q", original.ClientID, unmarshaled.ClientID)
	}
	if unmarshaled.ClientSecret != original.ClientSecret {
		t.Errorf("Expected client_secret %q, got %q", original.ClientSecret, unmarshaled.ClientSecret)
	}
	if unmarshaled.DeviceID != original.DeviceID {
		t.Errorf("Expected device_id %q, got %q", original.DeviceID, unmarshaled.DeviceID)
	}
	if unmarshaled.Username != original.Username {
		t.Errorf("Expected username %q, got %q", original.Username, unmarshaled.Username)
	}
	if unmarshaled.Password != original.Password {
		t.Errorf("Expected password %q, got %q", original.Password, unmarshaled.Password)
	}
}

func TestLoginResponse_JSONMarshalUnmarshal(t *testing.T) {
	original := &LoginResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "refresh-token-123",
		Scope:        "read write",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal login response: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled LoginResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	// Compare all fields
	if unmarshaled.AccessToken != original.AccessToken {
		t.Errorf("Expected access_token %q, got %q", original.AccessToken, unmarshaled.AccessToken)
	}
	if unmarshaled.TokenType != original.TokenType {
		t.Errorf("Expected token_type %q, got %q", original.TokenType, unmarshaled.TokenType)
	}
	if unmarshaled.ExpiresIn != original.ExpiresIn {
		t.Errorf("Expected expires_in %d, got %d", original.ExpiresIn, unmarshaled.ExpiresIn)
	}
	if unmarshaled.RefreshToken != original.RefreshToken {
		t.Errorf("Expected refresh_token %q, got %q", original.RefreshToken, unmarshaled.RefreshToken)
	}
	if unmarshaled.Scope != original.Scope {
		t.Errorf("Expected scope %q, got %q", original.Scope, unmarshaled.Scope)
	}
}

func TestLoginResponse_OptionalFields(t *testing.T) {
	// Test with minimal required fields
	minimalResponse := &LoginResponse{
		AccessToken: "token123",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	data, err := json.Marshal(minimalResponse)
	if err != nil {
		t.Fatalf("Failed to marshal minimal login response: %v", err)
	}

	var unmarshaled LoginResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal minimal login response: %v", err)
	}

	// Required fields should be preserved
	if unmarshaled.AccessToken != "token123" {
		t.Errorf("Expected access_token 'token123', got %q", unmarshaled.AccessToken)
	}
	if unmarshaled.TokenType != "Bearer" {
		t.Errorf("Expected token_type 'Bearer', got %q", unmarshaled.TokenType)
	}
	if unmarshaled.ExpiresIn != 3600 {
		t.Errorf("Expected expires_in 3600, got %d", unmarshaled.ExpiresIn)
	}

	// Optional fields should be empty
	if unmarshaled.RefreshToken != "" {
		t.Errorf("Expected empty refresh_token, got %q", unmarshaled.RefreshToken)
	}
	if unmarshaled.Scope != "" {
		t.Errorf("Expected empty scope, got %q", unmarshaled.Scope)
	}
}

func TestTokenInfo_JSONMarshalUnmarshal(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(time.Hour)

	original := &TokenInfo{
		Token:        "access-token-123",
		TokenType:    "Bearer",
		CreatedAt:    now,
		ExpiresIn:    3600,
		ExpiresAt:    expiresAt,
		Username:     "test@example.com",
		RefreshToken: "refresh-token-456",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal token info: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled TokenInfo
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal token info: %v", err)
	}

	// Compare all fields
	if unmarshaled.Token != original.Token {
		t.Errorf("Expected token %q, got %q", original.Token, unmarshaled.Token)
	}
	if unmarshaled.TokenType != original.TokenType {
		t.Errorf("Expected token_type %q, got %q", original.TokenType, unmarshaled.TokenType)
	}
	if unmarshaled.ExpiresIn != original.ExpiresIn {
		t.Errorf("Expected expires_in %d, got %d", original.ExpiresIn, unmarshaled.ExpiresIn)
	}
	if unmarshaled.Username != original.Username {
		t.Errorf("Expected username %q, got %q", original.Username, unmarshaled.Username)
	}
	if unmarshaled.RefreshToken != original.RefreshToken {
		t.Errorf("Expected refresh_token %q, got %q", original.RefreshToken, unmarshaled.RefreshToken)
	}

	// Check time fields (with some tolerance for serialization)
	timeDiff := unmarshaled.CreatedAt.Sub(original.CreatedAt)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("CreatedAt time mismatch: expected %v, got %v", original.CreatedAt, unmarshaled.CreatedAt)
	}

	timeDiff = unmarshaled.ExpiresAt.Sub(original.ExpiresAt)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("ExpiresAt time mismatch: expected %v, got %v", original.ExpiresAt, unmarshaled.ExpiresAt)
	}
}

func TestTokenInfo_TimeZoneHandling(t *testing.T) {
	// Test with different time zones
	utc := time.Now().UTC()
	local := time.Now()
	est, _ := time.LoadLocation("America/New_York")
	estTime := time.Now().In(est)

	testCases := []struct {
		name string
		time time.Time
	}{
		{"UTC time", utc},
		{"Local time", local},
		{"EST time", estTime},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenInfo := &TokenInfo{
				Token:     "test-token",
				CreatedAt: tc.time,
				ExpiresAt: tc.time.Add(time.Hour),
			}

			data, err := json.Marshal(tokenInfo)
			if err != nil {
				t.Fatalf("Failed to marshal token info with %s: %v", tc.name, err)
			}

			var unmarshaled TokenInfo
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal token info with %s: %v", tc.name, err)
			}

			// Times should be preserved (Go's JSON marshaling preserves timezone info)
			if !unmarshaled.CreatedAt.Equal(tc.time) {
				t.Errorf("CreatedAt not preserved: expected %v, got %v", tc.time, unmarshaled.CreatedAt)
			}

			expectedExpiresAt := tc.time.Add(time.Hour)
			if !unmarshaled.ExpiresAt.Equal(expectedExpiresAt) {
				t.Errorf("ExpiresAt not preserved: expected %v, got %v", expectedExpiresAt, unmarshaled.ExpiresAt)
			}
		})
	}
}

func TestUserProfile_JSONMarshalUnmarshal(t *testing.T) {
	original := &UserProfile{
		ID:       12345,
		Username: "testuser",
		Email:    "test@example.com",
		Name:     "Test User Full Name",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal user profile: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled UserProfile
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal user profile: %v", err)
	}

	// Compare all fields
	if unmarshaled.ID != original.ID {
		t.Errorf("Expected ID %d, got %d", original.ID, unmarshaled.ID)
	}
	if unmarshaled.Username != original.Username {
		t.Errorf("Expected username %q, got %q", original.Username, unmarshaled.Username)
	}
	if unmarshaled.Email != original.Email {
		t.Errorf("Expected email %q, got %q", original.Email, unmarshaled.Email)
	}
	if unmarshaled.Name != original.Name {
		t.Errorf("Expected name %q, got %q", original.Name, unmarshaled.Name)
	}
}

func TestLibraryResponse_JSONMarshalUnmarshal(t *testing.T) {
	lastOpenedAt := "2024-01-15T10:30:00Z"
	readProgress := 75
	readPercentage := 65.5

	original := &LibraryResponse{
		Summaries: []LibrarySummary{
			{
				ID:             123,
				Title:          "First Book",
				Description:    "Description of first book",
				LastOpenedAt:   &lastOpenedAt,
				ReadProgress:   &readProgress,
				ReadPercentage: &readPercentage,
			},
			{
				ID:             456,
				Title:          "Second Book",
				Description:    "Description of second book",
				LastOpenedAt:   nil,
				ReadProgress:   nil,
				ReadPercentage: nil,
			},
		},
		Total: 2,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal library response: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled LibraryResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal library response: %v", err)
	}

	// Compare total
	if unmarshaled.Total != original.Total {
		t.Errorf("Expected total %d, got %d", original.Total, unmarshaled.Total)
	}

	// Compare summaries count
	if len(unmarshaled.Summaries) != len(original.Summaries) {
		t.Errorf("Expected %d summaries, got %d", len(original.Summaries), len(unmarshaled.Summaries))
		return
	}

	// Compare first summary (with all fields)
	first := unmarshaled.Summaries[0]
	expectedFirst := original.Summaries[0]

	if first.ID != expectedFirst.ID {
		t.Errorf("Expected first summary ID %d, got %d", expectedFirst.ID, first.ID)
	}
	if first.Title != expectedFirst.Title {
		t.Errorf("Expected first summary title %q, got %q", expectedFirst.Title, first.Title)
	}
	if first.Description != expectedFirst.Description {
		t.Errorf("Expected first summary description %q, got %q", expectedFirst.Description, first.Description)
	}

	// Check pointer fields
	if first.LastOpenedAt == nil || *first.LastOpenedAt != *expectedFirst.LastOpenedAt {
		t.Errorf("Expected lastOpenedAt %v, got %v", expectedFirst.LastOpenedAt, first.LastOpenedAt)
	}
	if first.ReadProgress == nil || *first.ReadProgress != *expectedFirst.ReadProgress {
		t.Errorf("Expected readProgress %v, got %v", expectedFirst.ReadProgress, first.ReadProgress)
	}
	if first.ReadPercentage == nil || *first.ReadPercentage != *expectedFirst.ReadPercentage {
		t.Errorf("Expected readPercentage %v, got %v", expectedFirst.ReadPercentage, first.ReadPercentage)
	}

	// Compare second summary (with nil fields)
	second := unmarshaled.Summaries[1]
	expectedSecond := original.Summaries[1]

	if second.ID != expectedSecond.ID {
		t.Errorf("Expected second summary ID %d, got %d", expectedSecond.ID, second.ID)
	}
	if second.LastOpenedAt != nil {
		t.Errorf("Expected nil lastOpenedAt, got %v", second.LastOpenedAt)
	}
	if second.ReadProgress != nil {
		t.Errorf("Expected nil readProgress, got %v", second.ReadProgress)
	}
	if second.ReadPercentage != nil {
		t.Errorf("Expected nil readPercentage, got %v", second.ReadPercentage)
	}
}

func TestBookData_JSONMarshalUnmarshal(t *testing.T) {
	// Create complex nested data structures
	summary := map[string]interface{}{
		"id":          123,
		"title":       "Test Book",
		"description": "A test book for testing",
	}

	chapters := []interface{}{
		map[string]interface{}{
			"id":    1,
			"title": "Chapter 1",
		},
		map[string]interface{}{
			"id":    2,
			"title": "Chapter 2",
		},
	}

	content := map[string]interface{}{
		"documentId": "123",
		"body": map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "paragraph",
					"text": "This is test content",
				},
			},
		},
	}

	notes := []interface{}{
		map[string]interface{}{
			"id":    1,
			"note":  "Test note",
			"color": "yellow",
		},
	}

	original := &BookData{
		ID:       "123",
		Title:    "Test Book",
		Summary:  summary,
		Chapters: chapters,
		Content:  content,
		Notes:    notes,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal book data: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled BookData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal book data: %v", err)
	}

	// Compare basic fields
	if unmarshaled.ID != original.ID {
		t.Errorf("Expected ID %q, got %q", original.ID, unmarshaled.ID)
	}
	if unmarshaled.Title != original.Title {
		t.Errorf("Expected title %q, got %q", original.Title, unmarshaled.Title)
	}

	// Check that complex fields are preserved (as maps/slices)
	if unmarshaled.Summary == nil {
		t.Error("Expected summary to be preserved")
	}
	if unmarshaled.Chapters == nil {
		t.Error("Expected chapters to be preserved")
	}
	if unmarshaled.Content == nil {
		t.Error("Expected content to be preserved")
	}
	if unmarshaled.Notes == nil {
		t.Error("Expected notes to be preserved")
	}

	// Verify nested structure is preserved
	summaryMap := unmarshaled.Summary.(map[string]interface{}) //nolint:sloppyTypeAssert
	if title, exists := summaryMap["title"]; !exists || title != "Test Book" {
		t.Error("Expected summary title to be preserved")
	}

	chaptersSlice := unmarshaled.Chapters.([]interface{}) //nolint:gocritic
	if len(chaptersSlice) != 2 {
		t.Errorf("Expected 2 chapters, got %d", len(chaptersSlice))
	}
}

func TestAPIError_JSONMarshalUnmarshal(t *testing.T) {
	original := &APIError{
		StatusCode: 404,
		Message:    "Resource not found",
		Response:   `{"error": "not found", "code": 404}`,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal API error: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled APIError
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal API error: %v", err)
	}

	// Compare all fields
	if unmarshaled.StatusCode != original.StatusCode {
		t.Errorf("Expected status code %d, got %d", original.StatusCode, unmarshaled.StatusCode)
	}
	if unmarshaled.Message != original.Message {
		t.Errorf("Expected message %q, got %q", original.Message, unmarshaled.Message)
	}
	if unmarshaled.Response != original.Response {
		t.Errorf("Expected response %q, got %q", original.Response, unmarshaled.Response)
	}
}

func TestAPIError_ErrorMethod(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 500,
		Message:    "Internal server error occurred",
		Response:   `{"error": "server error"}`,
	}

	errorMsg := apiErr.Error()
	if errorMsg != apiErr.Message {
		t.Errorf("Expected Error() to return message %q, got %q", apiErr.Message, errorMsg)
	}
}

func TestAPIError_ErrorInterface(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 401,
		Message:    "Unauthorized access",
		Response:   `{"error": "unauthorized"}`,
	}

	// Test that APIError implements error interface
	var err error = apiErr
	if err.Error() != "Unauthorized access" {
		t.Errorf("Expected error message 'Unauthorized access', got %q", err.Error())
	}
}

func TestTypes_MalformedJSON(t *testing.T) {
	malformedJSON := `{"invalid": json, "missing": quote}`

	tests := []struct {
		name string
		ptr  interface{}
	}{
		{"Credentials", &Credentials{}},
		{"LoginRequest", &LoginRequest{}},
		{"LoginResponse", &LoginResponse{}},
		{"TokenInfo", &TokenInfo{}},
		{"UserProfile", &UserProfile{}},
		{"LibraryResponse", &LibraryResponse{}},
		{"BookData", &BookData{}},
		{"APIError", &APIError{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(malformedJSON), tt.ptr)
			if err == nil {
				t.Errorf("Expected error when unmarshaling malformed JSON for %s", tt.name)
			}
		})
	}
}

func TestTypes_EmptyJSON(t *testing.T) {
	emptyJSON := `{}`

	tests := []struct {
		name     string
		ptr      interface{}
		validate func(t *testing.T, obj interface{})
	}{
		{
			name: "Credentials",
			ptr:  &Credentials{},
			validate: func(t *testing.T, obj interface{}) {
				creds := obj.(*Credentials)
				if creds.Username != "" || creds.Password != "" {
					t.Error("Expected empty credentials")
				}
			},
		},
		{
			name: "UserProfile",
			ptr:  &UserProfile{},
			validate: func(t *testing.T, obj interface{}) {
				profile := obj.(*UserProfile)
				if profile.ID != 0 || profile.Username != "" || profile.Email != "" || profile.Name != "" {
					t.Error("Expected zero-value user profile")
				}
			},
		},
		{
			name: "LibraryResponse",
			ptr:  &LibraryResponse{},
			validate: func(t *testing.T, obj interface{}) {
				library := obj.(*LibraryResponse)
				if library.Total != 0 || len(library.Summaries) != 0 {
					t.Error("Expected zero-value library response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(emptyJSON), tt.ptr)
			if err != nil {
				t.Errorf("Unexpected error unmarshaling empty JSON for %s: %v", tt.name, err)
			}
			tt.validate(t, tt.ptr)
		})
	}
}

func TestTypes_NilPointers(t *testing.T) {
	// Test that pointer fields can handle nil values
	jsonWithNulls := `{
		"lastOpenedAt": null,
		"readProgress": null,
		"readPercentage": null,
		"refresh_token": null,
		"scope": null
	}`

	t.Run("LibrarySummary with nulls", func(t *testing.T) {
		var summary LibrarySummary
		err := json.Unmarshal([]byte(jsonWithNulls), &summary)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON with nulls: %v", err)
		}

		if summary.LastOpenedAt != nil {
			t.Error("Expected LastOpenedAt to be nil")
		}
		if summary.ReadProgress != nil {
			t.Error("Expected ReadProgress to be nil")
		}
		if summary.ReadPercentage != nil {
			t.Error("Expected ReadPercentage to be nil")
		}
	})

	t.Run("LoginResponse with nulls", func(t *testing.T) {
		var response LoginResponse
		err := json.Unmarshal([]byte(jsonWithNulls), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON with nulls: %v", err)
		}

		// Required fields should have zero values
		if response.AccessToken != "" {
			t.Error("Expected empty AccessToken")
		}
		if response.TokenType != "" {
			t.Error("Expected empty TokenType")
		}
		if response.ExpiresIn != 0 {
			t.Error("Expected zero ExpiresIn")
		}

		// Optional fields should be empty (not nil, but empty strings)
		if response.RefreshToken != "" {
			t.Error("Expected empty RefreshToken")
		}
		if response.Scope != "" {
			t.Error("Expected empty Scope")
		}
	})
}
