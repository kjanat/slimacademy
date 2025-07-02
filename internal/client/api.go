package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// extractStringFromMap safely extracts a string value from a map, handling type assertions
func extractStringFromMap(data any, key string) string {
	if dataMap, ok := data.(map[string]any); ok {
		if val, exists := dataMap[key]; exists {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	return ""
}

// GetProfile fetches the user profile
func (c *SlimClient) GetProfile(ctx context.Context) (*UserProfile, error) {
	req, err := c.newAuthenticatedRequest(ctx, "GET", "/api/user/profile", nil)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	var profile UserProfile
	if err := json.Unmarshal(respBody, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	return &profile, nil
}

// GetLibrary fetches the library summary (matching bash script)
func (c *SlimClient) GetLibrary(ctx context.Context) (*LibraryResponse, error) {
	endpoint := "/api/summary/library?sortBy=lastOpenedAt&sortOrder=DESC&onlyRecentlyRead=0"
	req, err := c.newAuthenticatedRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get library: %w", err)
	}

	var library LibraryResponse
	if err := json.Unmarshal(respBody, &library); err != nil {
		return nil, fmt.Errorf("failed to parse library response: %w", err)
	}

	return &library, nil
}

// GetSummary fetches summary data for a specific ID
func (c *SlimClient) GetSummary(ctx context.Context, id string) (any, error) {
	endpoint := fmt.Sprintf("/api/summary/%s", id)
	req, err := c.newAuthenticatedRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary for ID %s: %w", id, err)
	}

	var summary any
	if err := json.Unmarshal(respBody, &summary); err != nil {
		return nil, fmt.Errorf("failed to parse summary response: %w", err)
	}

	return summary, nil
}

// GetChapters fetches chapters data for a specific ID
func (c *SlimClient) GetChapters(ctx context.Context, id string) (any, error) {
	endpoint := fmt.Sprintf("/api/summary/%s/chapters", id)
	req, err := c.newAuthenticatedRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapters for ID %s: %w", id, err)
	}

	var chapters any
	if err := json.Unmarshal(respBody, &chapters); err != nil {
		return nil, fmt.Errorf("failed to parse chapters response: %w", err)
	}

	return chapters, nil
}

// GetContent fetches content data for a specific ID
func (c *SlimClient) GetContent(ctx context.Context, id string) (any, error) {
	endpoint := fmt.Sprintf("/api/summary/%s/content", id)
	req, err := c.newAuthenticatedRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get content for ID %s: %w", id, err)
	}

	var content any
	if err := json.Unmarshal(respBody, &content); err != nil {
		return nil, fmt.Errorf("failed to parse content response: %w", err)
	}

	return content, nil
}

// GetNotes fetches notes data for a specific ID
func (c *SlimClient) GetNotes(ctx context.Context, id string) (any, error) {
	endpoint := fmt.Sprintf("/api/summary/%s/list-notes", id)
	req, err := c.newAuthenticatedRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes for ID %s: %w", id, err)
	}

	var notes any
	if err := json.Unmarshal(respBody, &notes); err != nil {
		return nil, fmt.Errorf("failed to parse notes response: %w", err)
	}

	return notes, nil
}

// FetchAllBookData fetches all data for a specific book ID (matching bash script "all" command)
func (c *SlimClient) FetchAllBookData(ctx context.Context, id string) (*BookData, error) {
	// Ensure we're authenticated
	if err := c.EnsureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Fetch summary to get title
	summary, err := c.GetSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	// Extract title from summary
	title := extractStringFromMap(summary, "title")

	// Fetch all other data
	chapters, err := c.GetChapters(ctx, id)
	if err != nil {
		return nil, err
	}

	content, err := c.GetContent(ctx, id)
	if err != nil {
		return nil, err
	}

	notes, err := c.GetNotes(ctx, id)
	if err != nil {
		return nil, err
	}

	return &BookData{
		ID:       id,
		Title:    title,
		Summary:  summary,
		Chapters: chapters,
		Content:  content,
		Notes:    notes,
	}, nil
}

// FetchLibraryBooks fetches all books from the library
func (c *SlimClient) FetchLibraryBooks(ctx context.Context) ([]*BookData, error) {
	// Ensure we're authenticated
	if err := c.EnsureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Get library
	library, err := c.GetLibrary(ctx)
	if err != nil {
		return nil, err
	}

	var books []*BookData
	for _, summary := range library.Summaries {
		id := fmt.Sprintf("%d", summary.ID)
		bookData, err := c.FetchAllBookData(ctx, id)
		if err != nil {
			fmt.Printf("Warning: failed to fetch book %s: %v\n", id, err)
			continue
		}
		books = append(books, bookData)
	}

	return books, nil
}
