package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Test helper to create a client with a valid token
func createAuthenticatedClient(t *testing.T, serverURL string) *SlimClient {
	tempDir := createTempDir(t)
	client := NewSlimClient(tempDir)
	client.baseURL = serverURL

	// Create valid token
	tokenInfo := &TokenInfo{
		Token:     "test-access-token",
		TokenType: "Bearer",
		CreatedAt: time.Now(),
		ExpiresIn: 3600,
		ExpiresAt: time.Now().Add(time.Hour),
		Username:  "test@example.com",
	}

	if err := client.tokenStore.SaveToken(tokenInfo); err != nil {
		t.Fatalf("Failed to save test token: %v", err)
	}

	return client
}

// Test helper to verify authentication header
func verifyAuthHeader(t *testing.T, r *http.Request, expectedToken string) {
	auth := r.Header.Get("authorization")
	expected := "Bearer " + expectedToken
	if auth != expected {
		t.Errorf("Expected authorization header %q, got %q", expected, auth)
	}
}

// Test helper to safely encode JSON responses in mock servers
func encodeJSON(t *testing.T, w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		t.Errorf("Failed to encode JSON response: %v", err)
	}
}

// Test helper to safely write response data in mock servers
func writeResponse(t *testing.T, w http.ResponseWriter, data []byte) {
	if _, err := w.Write(data); err != nil {
		t.Errorf("Failed to write response: %v", err)
	}
}

func TestSlimClient_GetProfile(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse UserProfile
		serverStatus   int
		expectedError  bool
		errorContains  string
	}{
		{
			name: "successful profile fetch",
			serverResponse: UserProfile{
				ID:       123,
				Username: "testuser",
				Email:    "test@example.com",
				Name:     "Test User",
			},
			serverStatus:  http.StatusOK,
			expectedError: false,
		},
		{
			name:          "unauthorized",
			serverStatus:  http.StatusUnauthorized,
			expectedError: true,
			errorContains: "failed to get profile",
		},
		{
			name:          "server error",
			serverStatus:  http.StatusInternalServerError,
			expectedError: true,
			errorContains: "failed to get profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method and path
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				if r.URL.Path != "/api/user/profile" {
					t.Errorf("Expected path /api/user/profile, got %s", r.URL.Path)
				}

				// Verify authentication
				verifyAuthHeader(t, r, "test-access-token")

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					encodeJSON(t, w, tt.serverResponse)
				} else {
					writeResponse(t, w, []byte(`{"error": "request failed"}`))
				}
			}))
			defer server.Close()

			client := createAuthenticatedClient(t, server.URL)
			ctx := context.Background()

			profile, err := client.GetProfile(ctx)

			if (err != nil) != tt.expectedError {
				t.Errorf("GetProfile() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.expectedError {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %v", tt.errorContains, err)
				}
				return
			}

			if profile == nil {
				t.Error("Expected profile to be non-nil")
				return
			}

			// Compare profile fields
			if profile.ID != tt.serverResponse.ID {
				t.Errorf("Expected ID %d, got %d", tt.serverResponse.ID, profile.ID)
			}
			if profile.Username != tt.serverResponse.Username {
				t.Errorf("Expected username %q, got %q", tt.serverResponse.Username, profile.Username)
			}
			if profile.Email != tt.serverResponse.Email {
				t.Errorf("Expected email %q, got %q", tt.serverResponse.Email, profile.Email)
			}
			if profile.Name != tt.serverResponse.Name {
				t.Errorf("Expected name %q, got %q", tt.serverResponse.Name, profile.Name)
			}
		})
	}
}

func TestSlimClient_GetProfile_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeResponse(t, w, []byte(`invalid json`))
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	_, err := client.GetProfile(ctx)
	if err == nil {
		t.Error("Expected error with malformed JSON response")
	}

	if !strings.Contains(err.Error(), "failed to parse profile response") {
		t.Errorf("Expected parse error, got %v", err)
	}
}

func TestSlimClient_GetLibrary(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse LibraryResponse
		serverStatus   int
		expectedError  bool
		errorContains  string
	}{
		{
			name: "successful library fetch",
			serverResponse: LibraryResponse{
				Summaries: []LibrarySummary{
					{
						ID:             123,
						Title:          "Test Book 1",
						Description:    "First test book",
						LastOpenedAt:   stringPtr("2024-01-01T12:00:00Z"),
						ReadProgress:   intPtr(50),
						ReadPercentage: float64Ptr(75.5),
					},
					{
						ID:             456,
						Title:          "Test Book 2",
						Description:    "Second test book",
						LastOpenedAt:   nil,
						ReadProgress:   intPtr(0),
						ReadPercentage: float64Ptr(0.0),
					},
				},
				Total: 2,
			},
			serverStatus:  http.StatusOK,
			expectedError: false,
		},
		{
			name:          "unauthorized",
			serverStatus:  http.StatusUnauthorized,
			expectedError: true,
			errorContains: "failed to get library",
		},
		{
			name:          "server error",
			serverStatus:  http.StatusInternalServerError,
			expectedError: true,
			errorContains: "failed to get library",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method and path
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				if r.URL.Path != "/api/summary/library" {
					t.Errorf("Expected path /api/summary/library, got %s", r.URL.Path)
				}

				// Verify query parameters
				expectedQuery := "sortBy=lastOpenedAt&sortOrder=DESC&onlyRecentlyRead=0"
				if r.URL.RawQuery != expectedQuery {
					t.Errorf("Expected query %q, got %q", expectedQuery, r.URL.RawQuery)
				}

				// Verify authentication
				verifyAuthHeader(t, r, "test-access-token")

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					encodeJSON(t, w, tt.serverResponse)
				} else {
					writeResponse(t, w, []byte(`{"error": "request failed"}`))
				}
			}))
			defer server.Close()

			client := createAuthenticatedClient(t, server.URL)
			ctx := context.Background()

			library, err := client.GetLibrary(ctx)

			if (err != nil) != tt.expectedError {
				t.Errorf("GetLibrary() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.expectedError {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %v", tt.errorContains, err)
				}
				return
			}

			if library == nil {
				t.Error("Expected library to be non-nil")
				return
			}

			// Compare library fields
			if library.Total != tt.serverResponse.Total {
				t.Errorf("Expected total %d, got %d", tt.serverResponse.Total, library.Total)
			}

			if len(library.Summaries) != len(tt.serverResponse.Summaries) {
				t.Errorf("Expected %d summaries, got %d", len(tt.serverResponse.Summaries), len(library.Summaries))
				return
			}

			// Check first summary in detail
			if len(library.Summaries) > 0 {
				expected := tt.serverResponse.Summaries[0]
				actual := library.Summaries[0]

				if actual.ID != expected.ID {
					t.Errorf("Expected summary ID %d, got %d", expected.ID, actual.ID)
				}
				if actual.Title != expected.Title {
					t.Errorf("Expected title %q, got %q", expected.Title, actual.Title)
				}
				if actual.Description != expected.Description {
					t.Errorf("Expected description %q, got %q", expected.Description, actual.Description)
				}
			}
		})
	}
}

// Helper functions for pointer values
func stringPtr(s string) *string    { return &s }
func intPtr(i int) *int             { return &i }
func float64Ptr(f float64) *float64 { return &f }

func TestSlimClient_GetSummary(t *testing.T) {
	testID := "123"
	expectedResponse := map[string]interface{}{
		"id":          123,
		"title":       "Test Summary",
		"description": "Test description",
		"chapters":    []interface{}{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		expectedPath := "/api/summary/" + testID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify authentication
		verifyAuthHeader(t, r, "test-access-token")

		w.WriteHeader(http.StatusOK)
		encodeJSON(t, w, expectedResponse)
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	summary, err := client.GetSummary(ctx, testID)
	if err != nil {
		t.Fatalf("GetSummary() failed: %v", err)
	}

	if summary == nil {
		t.Fatal("Expected summary to be non-nil")
	}

	// Convert to map for comparison
	summaryMap := summary.(map[string]interface{}) //nolint:gocritic

	if title := summaryMap["title"].(string); title != "Test Summary" { //nolint:gocritic
		t.Errorf("Expected title 'Test Summary', got %q", title)
	}
}

func TestSlimClient_GetChapters(t *testing.T) {
	testID := "456"
	expectedResponse := []interface{}{
		map[string]interface{}{
			"id":    1,
			"title": "Chapter 1",
		},
		map[string]interface{}{
			"id":    2,
			"title": "Chapter 2",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		expectedPath := "/api/summary/" + testID + "/chapters"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify authentication
		verifyAuthHeader(t, r, "test-access-token")

		w.WriteHeader(http.StatusOK)
		encodeJSON(t, w, expectedResponse)
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	chapters, err := client.GetChapters(ctx, testID)
	if err != nil {
		t.Fatalf("GetChapters() failed: %v", err)
	}

	if chapters == nil {
		t.Fatal("Expected chapters to be non-nil")
	}

	// Convert to slice for comparison
	chaptersSlice := chapters.([]interface{}) //nolint:gocritic

	if len(chaptersSlice) != 2 {
		t.Errorf("Expected 2 chapters, got %d", len(chaptersSlice))
	}
}

func TestSlimClient_GetContent(t *testing.T) {
	testID := "789"
	expectedResponse := map[string]interface{}{
		"documentId": testID,
		"body": map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type":    "paragraph",
					"content": "Test content",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		expectedPath := "/api/summary/" + testID + "/content"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify authentication
		verifyAuthHeader(t, r, "test-access-token")

		w.WriteHeader(http.StatusOK)
		encodeJSON(t, w, expectedResponse)
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	content, err := client.GetContent(ctx, testID)
	if err != nil {
		t.Fatalf("GetContent() failed: %v", err)
	}

	if content == nil {
		t.Fatal("Expected content to be non-nil")
	}

	// Convert to map for comparison
	contentMap := content.(map[string]interface{}) //nolint:gocritic

	if docID := contentMap["documentId"].(string); docID != testID { //nolint:gocritic
		t.Errorf("Expected documentId %q, got %q", testID, docID)
	}
}

func TestSlimClient_GetNotes(t *testing.T) {
	testID := "101"
	expectedResponse := []interface{}{
		map[string]interface{}{
			"id":    1,
			"note":  "Test note 1",
			"color": "yellow",
		},
		map[string]interface{}{
			"id":    2,
			"note":  "Test note 2",
			"color": "blue",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		expectedPath := "/api/summary/" + testID + "/list-notes"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify authentication
		verifyAuthHeader(t, r, "test-access-token")

		w.WriteHeader(http.StatusOK)
		encodeJSON(t, w, expectedResponse)
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	notes, err := client.GetNotes(ctx, testID)
	if err != nil {
		t.Fatalf("GetNotes() failed: %v", err)
	}

	if notes == nil {
		t.Fatal("Expected notes to be non-nil")
	}

	// Convert to slice for comparison
	notesSlice := notes.([]interface{}) //nolint:gocritic

	if len(notesSlice) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notesSlice))
	}
}

func TestSlimClient_FetchAllBookData(t *testing.T) {
	testID := "999"

	// Mock responses for each endpoint
	summaryResp := map[string]interface{}{
		"id":    999,
		"title": "Complete Test Book",
	}
	chaptersResp := []interface{}{
		map[string]interface{}{"id": 1, "title": "Chapter 1"},
	}
	contentResp := map[string]interface{}{
		"documentId": testID,
		"body":       map[string]interface{}{"content": []interface{}{}},
	}
	notesResp := []interface{}{
		map[string]interface{}{"id": 1, "note": "Test note"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication for all requests
		verifyAuthHeader(t, r, "test-access-token")

		// Route to appropriate response based on path
		switch r.URL.Path {
		case "/api/summary/" + testID:
			encodeJSON(t, w, summaryResp)
		case "/api/summary/" + testID + "/chapters":
			encodeJSON(t, w, chaptersResp)
		case "/api/summary/" + testID + "/content":
			encodeJSON(t, w, contentResp)
		case "/api/summary/" + testID + "/list-notes":
			encodeJSON(t, w, notesResp)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	bookData, err := client.FetchAllBookData(ctx, testID)
	if err != nil {
		t.Fatalf("FetchAllBookData() failed: %v", err)
	}

	if bookData == nil {
		t.Fatal("Expected bookData to be non-nil")
	}

	// Verify book data fields
	if bookData.ID != testID {
		t.Errorf("Expected ID %q, got %q", testID, bookData.ID)
	}
	if bookData.Title != "Complete Test Book" {
		t.Errorf("Expected title 'Complete Test Book', got %q", bookData.Title)
	}
	if bookData.Summary == nil {
		t.Error("Expected summary to be non-nil")
	}
	if bookData.Chapters == nil {
		t.Error("Expected chapters to be non-nil")
	}
	if bookData.Content == nil {
		t.Error("Expected content to be non-nil")
	}
	if bookData.Notes == nil {
		t.Error("Expected notes to be non-nil")
	}
}

func TestSlimClient_FetchAllBookData_WithoutAuthentication(t *testing.T) {
	tempDir := createTempDir(t)

	// Create credentials for auto-login
	createCredentialsFile(t, tempDir, "test@example.com", "password123")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			// Mock login response
			resp := LoginResponse{
				AccessToken: "auto-login-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}
			encodeJSON(t, w, resp)
		} else if strings.HasPrefix(r.URL.Path, "/api/summary/") {
			// Mock API responses - verify we got the new token
			verifyAuthHeader(t, r, "auto-login-token")
			encodeJSON(t, w, map[string]interface{}{"id": 123, "title": "Auto Login Book"})
		}
	}))
	defer server.Close()

	client := NewSlimClient(tempDir)
	client.baseURL = server.URL
	client.credManager = NewCredentialManager(filepath.Join(tempDir, ".env"))

	ctx := context.Background()
	bookData, err := client.FetchAllBookData(ctx, "123")
	if err != nil {
		t.Fatalf("FetchAllBookData() with auto-login failed: %v", err)
	}

	if bookData.Title != "Auto Login Book" {
		t.Errorf("Expected title from auto-login, got %q", bookData.Title)
	}
}

func TestSlimClient_FetchLibraryBooks(t *testing.T) {
	libraryResp := LibraryResponse{
		Summaries: []LibrarySummary{
			{ID: 123, Title: "Book 1"},
			{ID: 456, Title: "Book 2"},
		},
		Total: 2,
	}

	// Track which book IDs were fetched
	fetchedBooks := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifyAuthHeader(t, r, "test-access-token")

		switch r.URL.Path {
		case "/api/summary/library":
			encodeJSON(t, w, libraryResp)
		case "/api/summary/123":
			fetchedBooks["123"] = true
			encodeJSON(t, w, map[string]interface{}{"id": 123, "title": "Book 1"})
		case "/api/summary/123/chapters":
			encodeJSON(t, w, []interface{}{})
		case "/api/summary/123/content":
			encodeJSON(t, w, map[string]interface{}{"documentId": "123"})
		case "/api/summary/123/list-notes":
			encodeJSON(t, w, []interface{}{})
		case "/api/summary/456":
			fetchedBooks["456"] = true
			encodeJSON(t, w, map[string]interface{}{"id": 456, "title": "Book 2"})
		case "/api/summary/456/chapters":
			encodeJSON(t, w, []interface{}{})
		case "/api/summary/456/content":
			encodeJSON(t, w, map[string]interface{}{"documentId": "456"})
		case "/api/summary/456/list-notes":
			encodeJSON(t, w, []interface{}{})
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	books, err := client.FetchLibraryBooks(ctx)
	if err != nil {
		t.Fatalf("FetchLibraryBooks() failed: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(books))
	}

	// Verify both books were fetched
	if !fetchedBooks["123"] {
		t.Error("Expected book 123 to be fetched")
	}
	if !fetchedBooks["456"] {
		t.Error("Expected book 456 to be fetched")
	}

	// Verify book data
	for i, book := range books {
		expectedID := strconv.Itoa(libraryResp.Summaries[i].ID)
		if book.ID != expectedID {
			t.Errorf("Expected book ID %q, got %q", expectedID, book.ID)
		}
		if book.Title != libraryResp.Summaries[i].Title {
			t.Errorf("Expected book title %q, got %q", libraryResp.Summaries[i].Title, book.Title)
		}
	}
}

func TestSlimClient_FetchLibraryBooks_PartialFailure(t *testing.T) {
	libraryResp := LibraryResponse{
		Summaries: []LibrarySummary{
			{ID: 123, Title: "Good Book"},
			{ID: 456, Title: "Bad Book"}, // This one will fail
		},
		Total: 2,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifyAuthHeader(t, r, "test-access-token")

		switch r.URL.Path {
		case "/api/summary/library":
			encodeJSON(t, w, libraryResp)
		case "/api/summary/123":
			encodeJSON(t, w, map[string]interface{}{"id": 123, "title": "Good Book"})
		case "/api/summary/123/chapters":
			encodeJSON(t, w, []interface{}{})
		case "/api/summary/123/content":
			encodeJSON(t, w, map[string]interface{}{"documentId": "123"})
		case "/api/summary/123/list-notes":
			encodeJSON(t, w, []interface{}{})
		case "/api/summary/456":
			// Simulate failure for book 456
			w.WriteHeader(http.StatusInternalServerError)
			writeResponse(t, w, []byte(`{"error": "internal server error"}`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	books, err := client.FetchLibraryBooks(ctx)
	if err != nil {
		t.Fatalf("FetchLibraryBooks() failed: %v", err)
	}

	// Should only get the successful book
	if len(books) != 1 {
		t.Errorf("Expected 1 successful book, got %d", len(books))
	}

	if len(books) > 0 && books[0].Title != "Good Book" {
		t.Errorf("Expected successful book to be 'Good Book', got %q", books[0].Title)
	}
}

func TestSlimClient_API_AuthenticationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return unauthorized
		w.WriteHeader(http.StatusUnauthorized)
		writeResponse(t, w, []byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)
	ctx := context.Background()

	// Test all API methods fail with authentication error
	_, err := client.GetProfile(ctx)
	if err == nil {
		t.Error("Expected GetProfile to fail with authentication error")
	}

	_, err = client.GetLibrary(ctx)
	if err == nil {
		t.Error("Expected GetLibrary to fail with authentication error")
	}

	_, err = client.GetSummary(ctx, "123")
	if err == nil {
		t.Error("Expected GetSummary to fail with authentication error")
	}

	_, err = client.GetChapters(ctx, "123")
	if err == nil {
		t.Error("Expected GetChapters to fail with authentication error")
	}

	_, err = client.GetContent(ctx, "123")
	if err == nil {
		t.Error("Expected GetContent to fail with authentication error")
	}

	_, err = client.GetNotes(ctx, "123")
	if err == nil {
		t.Error("Expected GetNotes to fail with authentication error")
	}
}

func TestSlimClient_API_MalformedResponses(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		apiCall  func(*SlimClient, context.Context) error
	}{
		{
			name:     "GetProfile malformed",
			endpoint: "/api/user/profile",
			apiCall: func(c *SlimClient, ctx context.Context) error {
				_, err := c.GetProfile(ctx)
				return err
			},
		},
		{
			name:     "GetLibrary malformed",
			endpoint: "/api/summary/library",
			apiCall: func(c *SlimClient, ctx context.Context) error {
				_, err := c.GetLibrary(ctx)
				return err
			},
		},
		{
			name:     "GetSummary malformed",
			endpoint: "/api/summary/123",
			apiCall: func(c *SlimClient, ctx context.Context) error {
				_, err := c.GetSummary(ctx, "123")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tt.endpoint {
					w.WriteHeader(http.StatusOK)
					writeResponse(t, w, []byte(`invalid json response`))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			client := createAuthenticatedClient(t, server.URL)
			ctx := context.Background()

			err := tt.apiCall(client, ctx)
			if err == nil {
				t.Errorf("Expected error with malformed JSON for %s", tt.name)
			}

			if !strings.Contains(err.Error(), "failed to parse") {
				t.Errorf("Expected parse error for %s, got %v", tt.name, err)
			}
		})
	}
}

func TestSlimClient_extractStringFromMap(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		key      string
		expected string
	}{
		{
			name:     "valid string value",
			data:     map[string]any{"title": "Test Title"},
			key:      "title",
			expected: "Test Title",
		},
		{
			name:     "missing key",
			data:     map[string]any{"other": "value"},
			key:      "title",
			expected: "",
		},
		{
			name:     "non-string value",
			data:     map[string]any{"title": 123},
			key:      "title",
			expected: "",
		},
		{
			name:     "non-map data",
			data:     "not a map",
			key:      "title",
			expected: "",
		},
		{
			name:     "nil data",
			data:     nil,
			key:      "title",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractStringFromMap(tt.data, tt.key)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSlimClient_ContextCancellation(t *testing.T) {
	// Create a server that responds slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		encodeJSON(t, w, map[string]interface{}{"id": 123})
	}))
	defer server.Close()

	client := createAuthenticatedClient(t, server.URL)

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetProfile(ctx)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected context deadline error, got %v", err)
	}
}
