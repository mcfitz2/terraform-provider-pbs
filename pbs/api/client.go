/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package api provides the HTTP client for the Proxmox Backup Server API
package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Client represents a PBS API client
type Client struct {
	httpClient    *http.Client
	endpoint      string
	apiToken      string
	username      string
	password      string
	ticket        string
	csrfToken     string
	authenticated bool
}

// Credentials holds authentication information
type Credentials struct {
	Username string
	Password string
	APIToken string
}

// ClientOptions holds configuration options for the client
type ClientOptions struct {
	Endpoint string
	Insecure bool
	Timeout  time.Duration
}

// NewClient creates a new PBS API client
func NewClient(creds Credentials, opts ClientOptions) (*Client, error) {
	if opts.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	if creds.APIToken == "" && (creds.Username == "" || creds.Password == "") {
		return nil, fmt.Errorf("either API token or username/password is required")
	}

	// Parse and validate endpoint URL
	u, err := url.Parse(opts.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("endpoint must use http or https scheme")
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.Insecure, //nolint:gosec
		},
	}

	if opts.Insecure {
		fmt.Println("[WARNING] TLS certificate verification is disabled (Insecure=true). This should only be used for development or testing environments.")
	}

	client := &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
		endpoint:      strings.TrimSuffix(opts.Endpoint, "/"),
		apiToken:      creds.APIToken,
		username:      creds.Username,
		password:      creds.Password,
		authenticated: false,
	}

	// If using username/password, authenticate immediately
	if client.apiToken == "" && client.username != "" && client.password != "" {
		if err := client.authenticate(); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	return client, nil
}

// APIResponse represents a standard PBS API response
type APIResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors interface{}     `json:"errors,omitempty"`
}

// AuthResponse represents a PBS authentication response
type AuthResponse struct {
	Data struct {
		Ticket    string `json:"ticket"`
		CSRFToken string `json:"CSRFPreventionToken"`
	} `json:"data"`
}

// DoRequest performs an HTTP request to the PBS API
func (c *Client) DoRequest(ctx context.Context, method, apiPath string, body interface{}) (*APIResponse, error) {
	u := fmt.Sprintf("%s/api2/json%s", c.endpoint, apiPath)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set authentication
	if c.apiToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("PBSAPIToken=%s", c.apiToken))
	} else if c.authenticated && c.ticket != "" {
		// Use ticket-based authentication
		req.Header.Set("Cookie", fmt.Sprintf("PBSAuthCookie=%s", c.ticket))
		if method != "GET" && c.csrfToken != "" {
			req.Header.Set("CSRFPreventionToken", c.csrfToken)
		}
	} else {
		return nil, fmt.Errorf("authentication failed - no authentication credentials provided")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Errors != nil {
		return nil, fmt.Errorf("API returned errors: %v", apiResp.Errors)
	}

	return &apiResp, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string) (*APIResponse, error) {
	return c.DoRequest(ctx, "GET", path, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*APIResponse, error) {
	return c.DoRequest(ctx, "POST", path, body)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*APIResponse, error) {
	return c.DoRequest(ctx, "PUT", path, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*APIResponse, error) {
	return c.DoRequest(ctx, "DELETE", path, nil)
}

// authenticate performs username/password authentication to get a ticket
func (c *Client) authenticate() error {
	loginData := url.Values{
		"username": {c.username},
		"password": {c.password},
	}

	u := fmt.Sprintf("%s/api2/json/access/ticket", c.endpoint)

	resp, err := c.httpClient.PostForm(u, loginData)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if authResp.Data.Ticket == "" {
		return fmt.Errorf("login successful but no ticket received")
	}

	c.ticket = authResp.Data.Ticket
	c.csrfToken = authResp.Data.CSRFToken
	c.authenticated = true

	return nil
}

// NodeInfo represents PBS node information
type NodeInfo struct {
	Node   string `json:"node"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

// TaskStatus represents PBS task status
type TaskStatus struct {
	Status    string      `json:"status"`               // "running", "stopped", etc.
	ExitCode  interface{} `json:"exitstatus,omitempty"` // Can be string or int
	StartTime int64       `json:"starttime"`
	EndTime   *int64      `json:"endtime,omitempty"`
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	Node      string      `json:"node"`
}

// GetExitCode safely extracts the exit code as an integer
func (ts *TaskStatus) GetExitCode() int {
	// Parse exit code from PBS task status

	if ts.ExitCode == nil {
		// No exit code provided
		// If task is stopped and no exit code, assume success
		if ts.Status == "stopped" {
			return 0
		}
		return -1 // Unknown exit code
	}

	switch v := ts.ExitCode.(type) {
	case int:
		// Exit code as integer
		return v
	case float64:
		code := int(v)
		// Exit code as float64, converted to int
		return code
	case string:
		// Exit code as string
		// Try to parse string as int
		if v == "" {
			// Empty string for stopped task likely means success
			if ts.Status == "stopped" {
				return 0
			}
			return -1
		}
		// For PBS, "OK" often means success (exit code 0)
		if v == "OK" {
			return 0
		}
		// Try to parse as number
		var code int
		if _, err := fmt.Sscanf(v, "%d", &code); err == nil {
			return code
		}
		return -1
	default:
		// Unknown exit code type
		return -1
	}
}

// GetNodes retrieves the list of PBS nodes
func (c *Client) GetNodes(ctx context.Context) ([]NodeInfo, error) {
	resp, err := c.Get(ctx, "/nodes")
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var nodes []NodeInfo
	if err := json.Unmarshal(resp.Data, &nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes response: %w", err)
	}

	return nodes, nil
}

// GetTaskStatus retrieves the status of a specific task
func (c *Client) GetTaskStatus(ctx context.Context, node, upid string) (*TaskStatus, error) {
	path := fmt.Sprintf("/nodes/%s/tasks/%s/status", url.PathEscape(node), url.PathEscape(upid))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get task status: %w", err)
	}

	var status TaskStatus
	if err := json.Unmarshal(resp.Data, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task status: %w", err)
	}

	return &status, nil
}

// WaitForTask waits for a PBS task to complete with timeout
func (c *Client) WaitForTask(ctx context.Context, node, upid string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	startTime := time.Now()
	lastLog := startTime

	for time.Now().Before(deadline) {
		status, err := c.GetTaskStatus(ctx, node, upid)
		if err != nil {
			return fmt.Errorf("failed to check task status: %w", err)
		}

		switch status.Status {
		case "stopped":
			exitCode := status.GetExitCode()
			if exitCode == 0 {
				return nil // Task completed successfully
			}
			// Task failed - return detailed error
			elapsed := time.Since(startTime).Round(time.Second)
			return fmt.Errorf("task failed with exit code %d after %v", exitCode, elapsed)
		case "running":
			// Log progress every 10 seconds for long-running tasks (visible with TF_LOG=DEBUG)
			if time.Since(lastLog) >= 10*time.Second {
				lastLog = time.Now()
				// Task still running, will continue polling
			}
			// Continue waiting
			time.Sleep(2 * time.Second)
		default:
			return fmt.Errorf("unknown task status: %s", status.Status)
		}
	}

	elapsed := time.Since(startTime).Round(time.Second)
	return fmt.Errorf("task did not complete within timeout %v (elapsed: %v, UPID: %s)", timeout, elapsed, upid)
}

// BuildPath safely constructs API paths
func BuildPath(segments ...string) string {
	cleanSegments := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment != "" {
			cleanSegments = append(cleanSegments, url.PathEscape(segment))
		}
	}
	return "/" + path.Join(cleanSegments...)
}
