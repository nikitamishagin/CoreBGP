package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nikitamishagin/corebgp/internal/model"
)

// APIClient represents the client for interacting with the API server.
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new API client instance.
func NewAPIClient(baseURL *string, timeout time.Duration) *APIClient {
	return &APIClient{
		baseURL: *baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// V1HealthCheck checks the health status of the API server (Version 1).
func (c *APIClient) V1HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/healthz", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status code %d", resp.StatusCode)
	}

	return nil
}

// V1GetAnnouncement retrieves an announcement by project and name.
func (c *APIClient) V1GetAnnouncement(ctx context.Context, project, name string) (*model.Announcement, error) {
	url := fmt.Sprintf("%s/v1/announcements/%s/%s", c.baseURL, project, name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("announcement not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch announcement: status code %d", resp.StatusCode)
	}

	var announcement model.Announcement
	if err := json.NewDecoder(resp.Body).Decode(&announcement); err != nil {
		return nil, fmt.Errorf("failed to decode announcement: %v", err)
	}

	return &announcement, nil
}

// V1CreateAnnouncement creates a new announcement.
func (c *APIClient) V1CreateAnnouncement(ctx context.Context, announcement *model.Announcement) error {
	url := c.baseURL + "/v1/announcements/"

	data, err := json.Marshal(announcement)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("announcement already exists")
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create announcement: status code %d", resp.StatusCode)
	}

	return nil
}

// V1UpdateAnnouncement updates an existing announcement.
func (c *APIClient) V1UpdateAnnouncement(ctx context.Context, announcement *model.Announcement) error {
	url := c.baseURL + "/v1/announcements/"

	data, err := json.Marshal(announcement)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("announcement not found")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update announcement: status code %d", resp.StatusCode)
	}

	return nil
}

// V1DeleteAnnouncement deletes an announcement by project and name.
func (c *APIClient) V1DeleteAnnouncement(ctx context.Context, project, name string) error {
	url := fmt.Sprintf("%s/v1/announcements/%s/%s", c.baseURL, project, name)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("announcement not found")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete announcement: status code %d", resp.StatusCode)
	}

	return nil
}

// V1WatchAnnouncements establishes a WebSocket connection to watch announcements.
func (c *APIClient) V1WatchAnnouncements(ctx context.Context, onEvent func(event map[string]interface{})) error {
	url := fmt.Sprintf("ws://%s/v1/watch/announcements/", c.baseURL)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to establish websocket connection: %v", err)
	}
	defer conn.Close()

	done := make(chan struct{})

	// Goroutine to read events from WebSocket.
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var event map[string]interface{}
			if err := json.Unmarshal(message, &event); err != nil {
				fmt.Printf("failed to unmarshal websocket message: %v\n", err)
				continue
			}

			onEvent(event)
		}
	}()

	<-done
	return nil
}
