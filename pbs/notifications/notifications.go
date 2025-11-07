/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package notifications provides API client functionality for PBS notification target configurations
package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// Client represents the notifications API client
type Client struct {
	api *api.Client
}

// NewClient creates a new notifications API client
func NewClient(apiClient *api.Client) *Client {
	return &Client{api: apiClient}
}

// NotificationTargetType represents the type of notification target
type NotificationTargetType string

const (
	NotificationTargetTypeSMTP     NotificationTargetType = "smtp"
	NotificationTargetTypeGotify   NotificationTargetType = "gotify"
	NotificationTargetTypeSendmail NotificationTargetType = "sendmail"
	NotificationTargetTypeWebhook  NotificationTargetType = "webhook"
)

// SMTPTarget represents an SMTP notification target configuration
type SMTPTarget struct {
	Name       string   `json:"name"`
	From       string   `json:"from-address"`
	To         []string `json:"mailto"` // PBS 4.0: array of email addresses
	Server     string   `json:"server"`
	Port       *int     `json:"port,omitempty"`
	Mode       string   `json:"mode,omitempty"` // insecure, starttls, tls
	Username   string   `json:"username,omitempty"`
	Password   string   `json:"password,omitempty"`
	Author     string   `json:"author,omitempty"`
	Comment    string   `json:"comment,omitempty"`
	Disable    *bool    `json:"disable,omitempty"`
	MailtoUser []string `json:"mailto-user,omitempty"`
	Origin     string   `json:"origin,omitempty"`
}

// GotifyTarget represents a Gotify notification target configuration
type GotifyTarget struct {
	Name    string `json:"name"`
	Server  string `json:"server"`
	Token   string `json:"token"`
	Comment string `json:"comment,omitempty"`
	Disable *bool  `json:"disable,omitempty"`
	Origin  string `json:"origin,omitempty"`
}

// SendmailTarget represents a Sendmail notification target configuration
type SendmailTarget struct {
	Name       string   `json:"name"`
	From       string   `json:"from-address"`
	Mailto     []string `json:"mailto,omitempty"` // PBS 4.0: array of email addresses
	MailtoUser []string `json:"mailto-user,omitempty"`
	Author     string   `json:"author,omitempty"`
	Comment    string   `json:"comment,omitempty"`
	Disable    *bool    `json:"disable,omitempty"`
	Origin     string   `json:"origin,omitempty"`
}

// WebhookTarget represents a Webhook notification target configuration
type WebhookTarget struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Body    string            `json:"body,omitempty"`
	Method  string            `json:"method,omitempty"` // POST, PUT
	Headers map[string]string `json:"header,omitempty"`
	Secret  string            `json:"secret,omitempty"`
	Comment string            `json:"comment,omitempty"`
	Disable *bool             `json:"disable,omitempty"`
	Origin  string            `json:"origin,omitempty"`
}

// SMTP Target Methods

// ListSMTPTargets lists all SMTP notification target configurations
func (c *Client) ListSMTPTargets(ctx context.Context) ([]SMTPTarget, error) {
	resp, err := c.api.Get(ctx, "/config/notifications/endpoints/smtp")
	if err != nil {
		return nil, fmt.Errorf("failed to list SMTP targets: %w", err)
	}

	var targets []SMTPTarget
	if err := json.Unmarshal(resp.Data, &targets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SMTP targets: %w", err)
	}

	return targets, nil
}

// GetSMTPTarget gets a specific SMTP notification target by name
func (c *Client) GetSMTPTarget(ctx context.Context, name string) (*SMTPTarget, error) {
	path := fmt.Sprintf("/config/notifications/endpoints/smtp/%s", url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMTP target %s: %w", name, err)
	}

	var target SMTPTarget
	if err := json.Unmarshal(resp.Data, &target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SMTP target %s: %w", name, err)
	}

	return &target, nil
}

// CreateSMTPTarget creates a new SMTP notification target
func (c *Client) CreateSMTPTarget(ctx context.Context, target *SMTPTarget) error {
	if target.Name == "" {
		return fmt.Errorf("target name is required")
	}
	if target.Server == "" {
		return fmt.Errorf("server is required")
	}
	if target.From == "" {
		return fmt.Errorf("from address is required")
	}

	body := map[string]interface{}{
		"name":         target.Name,
		"server":       target.Server,
		"from-address": target.From,
	}

	if target.To != nil {
		body["mailto"] = target.To
	}
	if target.MailtoUser != nil {
		body["mailto-user"] = target.MailtoUser
	}
	if target.Port != nil {
		body["port"] = *target.Port
	}
	if target.Mode != "" {
		body["mode"] = target.Mode
	}
	if target.Username != "" {
		body["username"] = target.Username
	}
	if target.Password != "" {
		body["password"] = target.Password
	}
	if target.Author != "" {
		body["author"] = target.Author
	}
	if target.Comment != "" {
		body["comment"] = target.Comment
	}
	if target.Disable != nil {
		body["disable"] = *target.Disable
	}

	_, err := c.api.Post(ctx, "/config/notifications/endpoints/smtp", body)
	if err != nil {
		return fmt.Errorf("failed to create SMTP target %s: %w", target.Name, err)
	}

	return nil
}

// UpdateSMTPTarget updates an existing SMTP notification target
func (c *Client) UpdateSMTPTarget(ctx context.Context, name string, target *SMTPTarget) error {
	if name == "" {
		return fmt.Errorf("target name is required")
	}

	body := map[string]interface{}{}

	if target.Server != "" {
		body["server"] = target.Server
	}
	if target.From != "" {
		body["from-address"] = target.From
	}
	if target.To != nil {
		body["mailto"] = target.To
	}
	if target.MailtoUser != nil {
		body["mailto-user"] = target.MailtoUser
	}
	if target.Port != nil {
		body["port"] = *target.Port
	}
	if target.Mode != "" {
		body["mode"] = target.Mode
	}
	if target.Username != "" {
		body["username"] = target.Username
	}
	if target.Password != "" {
		body["password"] = target.Password
	}
	if target.Author != "" {
		body["author"] = target.Author
	}
	if target.Comment != "" {
		body["comment"] = target.Comment
	}
	if target.Disable != nil {
		body["disable"] = *target.Disable
	}

	path := fmt.Sprintf("/config/notifications/endpoints/smtp/%s", url.PathEscape(name))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update SMTP target %s: %w", name, err)
	}

	return nil
}

// DeleteSMTPTarget deletes an SMTP notification target
func (c *Client) DeleteSMTPTarget(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("target name is required")
	}

	path := fmt.Sprintf("/config/notifications/endpoints/smtp/%s", url.PathEscape(name))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete SMTP target %s: %w", name, err)
	}

	return nil
}

// Gotify Target Methods

// ListGotifyTargets lists all Gotify notification target configurations
func (c *Client) ListGotifyTargets(ctx context.Context) ([]GotifyTarget, error) {
	resp, err := c.api.Get(ctx, "/config/notifications/endpoints/gotify")
	if err != nil {
		return nil, fmt.Errorf("failed to list Gotify targets: %w", err)
	}

	var targets []GotifyTarget
	if err := json.Unmarshal(resp.Data, &targets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gotify targets: %w", err)
	}

	return targets, nil
}

// GetGotifyTarget gets a specific Gotify notification target by name
func (c *Client) GetGotifyTarget(ctx context.Context, name string) (*GotifyTarget, error) {
	path := fmt.Sprintf("/config/notifications/endpoints/gotify/%s", url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get Gotify target %s: %w", name, err)
	}

	var target GotifyTarget
	if err := json.Unmarshal(resp.Data, &target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gotify target %s: %w", name, err)
	}

	return &target, nil
}

// CreateGotifyTarget creates a new Gotify notification target
func (c *Client) CreateGotifyTarget(ctx context.Context, target *GotifyTarget) error {
	if target.Name == "" {
		return fmt.Errorf("target name is required")
	}
	if target.Server == "" {
		return fmt.Errorf("server is required")
	}
	if target.Token == "" {
		return fmt.Errorf("token is required")
	}

	body := map[string]interface{}{
		"name":   target.Name,
		"server": target.Server,
		"token":  target.Token,
	}

	if target.Comment != "" {
		body["comment"] = target.Comment
	}
	if target.Disable != nil {
		body["disable"] = *target.Disable
	}

	_, err := c.api.Post(ctx, "/config/notifications/endpoints/gotify", body)
	if err != nil {
		return fmt.Errorf("failed to create Gotify target %s: %w", target.Name, err)
	}

	return nil
}

// UpdateGotifyTarget updates an existing Gotify notification target
func (c *Client) UpdateGotifyTarget(ctx context.Context, name string, target *GotifyTarget) error {
	if name == "" {
		return fmt.Errorf("target name is required")
	}

	body := map[string]interface{}{}

	if target.Server != "" {
		body["server"] = target.Server
	}
	if target.Token != "" {
		body["token"] = target.Token
	}
	if target.Comment != "" {
		body["comment"] = target.Comment
	}
	if target.Disable != nil {
		body["disable"] = *target.Disable
	}

	path := fmt.Sprintf("/config/notifications/endpoints/gotify/%s", url.PathEscape(name))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update Gotify target %s: %w", name, err)
	}

	return nil
}

// DeleteGotifyTarget deletes a Gotify notification target
func (c *Client) DeleteGotifyTarget(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("target name is required")
	}

	path := fmt.Sprintf("/config/notifications/endpoints/gotify/%s", url.PathEscape(name))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete Gotify target %s: %w", name, err)
	}

	return nil
}

// Sendmail Target Methods

// ListSendmailTargets lists all Sendmail notification target configurations
func (c *Client) ListSendmailTargets(ctx context.Context) ([]SendmailTarget, error) {
	resp, err := c.api.Get(ctx, "/config/notifications/endpoints/sendmail")
	if err != nil {
		return nil, fmt.Errorf("failed to list Sendmail targets: %w", err)
	}

	var targets []SendmailTarget
	if err := json.Unmarshal(resp.Data, &targets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Sendmail targets: %w", err)
	}

	return targets, nil
}

// GetSendmailTarget gets a specific Sendmail notification target by name
func (c *Client) GetSendmailTarget(ctx context.Context, name string) (*SendmailTarget, error) {
	path := fmt.Sprintf("/config/notifications/endpoints/sendmail/%s", url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get Sendmail target %s: %w", name, err)
	}

	var target SendmailTarget
	if err := json.Unmarshal(resp.Data, &target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Sendmail target %s: %w", name, err)
	}

	return &target, nil
}

// CreateSendmailTarget creates a new Sendmail notification target
func (c *Client) CreateSendmailTarget(ctx context.Context, target *SendmailTarget) error {
	if target.Name == "" {
		return fmt.Errorf("target name is required")
	}
	if target.From == "" {
		return fmt.Errorf("from address is required")
	}

	body := map[string]interface{}{
		"name":         target.Name,
		"from-address": target.From,
	}

	if target.Mailto != nil {
		body["mailto"] = target.Mailto
	}
	if target.MailtoUser != nil {
		body["mailto-user"] = target.MailtoUser
	}
	if target.Author != "" {
		body["author"] = target.Author
	}
	if target.Comment != "" {
		body["comment"] = target.Comment
	}
	if target.Disable != nil {
		body["disable"] = *target.Disable
	}

	_, err := c.api.Post(ctx, "/config/notifications/endpoints/sendmail", body)
	if err != nil {
		return fmt.Errorf("failed to create Sendmail target %s: %w", target.Name, err)
	}

	return nil
}

// UpdateSendmailTarget updates an existing Sendmail notification target
func (c *Client) UpdateSendmailTarget(ctx context.Context, name string, target *SendmailTarget) error {
	if name == "" {
		return fmt.Errorf("target name is required")
	}

	body := map[string]interface{}{}

	if target.From != "" {
		body["from-address"] = target.From
	}
	if target.Mailto != nil {
		body["mailto"] = target.Mailto
	}
	if target.MailtoUser != nil {
		body["mailto-user"] = target.MailtoUser
	}
	if target.Author != "" {
		body["author"] = target.Author
	}
	if target.Comment != "" {
		body["comment"] = target.Comment
	}
	if target.Disable != nil {
		body["disable"] = *target.Disable
	}

	path := fmt.Sprintf("/config/notifications/endpoints/sendmail/%s", url.PathEscape(name))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update Sendmail target %s: %w", name, err)
	}

	return nil
}

// DeleteSendmailTarget deletes a Sendmail notification target
func (c *Client) DeleteSendmailTarget(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("target name is required")
	}

	path := fmt.Sprintf("/config/notifications/endpoints/sendmail/%s", url.PathEscape(name))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete Sendmail target %s: %w", name, err)
	}

	return nil
}

// Webhook Target Methods

// ListWebhookTargets lists all Webhook notification target configurations
func (c *Client) ListWebhookTargets(ctx context.Context) ([]WebhookTarget, error) {
	resp, err := c.api.Get(ctx, "/config/notifications/endpoints/webhook")
	if err != nil {
		return nil, fmt.Errorf("failed to list Webhook targets: %w", err)
	}

	var targets []WebhookTarget
	if err := json.Unmarshal(resp.Data, &targets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Webhook targets: %w", err)
	}

	return targets, nil
}

// GetWebhookTarget gets a specific Webhook notification target by name
func (c *Client) GetWebhookTarget(ctx context.Context, name string) (*WebhookTarget, error) {
	path := fmt.Sprintf("/config/notifications/endpoints/webhook/%s", url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get Webhook target %s: %w", name, err)
	}

	var target WebhookTarget
	if err := json.Unmarshal(resp.Data, &target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Webhook target %s: %w", name, err)
	}

	return &target, nil
}

// CreateWebhookTarget creates a new Webhook notification target
func (c *Client) CreateWebhookTarget(ctx context.Context, target *WebhookTarget) error {
	if target.Name == "" {
		return fmt.Errorf("target name is required")
	}
	if target.URL == "" {
		return fmt.Errorf("url is required")
	}

	body := map[string]interface{}{
		"name": target.Name,
		"url":  target.URL,
	}

	if target.Body != "" {
		body["body"] = target.Body
	}
	if target.Method != "" {
		body["method"] = target.Method
	}
	if target.Secret != "" {
		body["secret"] = target.Secret
	}
	if len(target.Headers) > 0 {
		body["header"] = target.Headers
	}
	if target.Comment != "" {
		body["comment"] = target.Comment
	}
	if target.Disable != nil {
		body["disable"] = *target.Disable
	}

	_, err := c.api.Post(ctx, "/config/notifications/endpoints/webhook", body)
	if err != nil {
		return fmt.Errorf("failed to create Webhook target %s: %w", target.Name, err)
	}

	return nil
}

// UpdateWebhookTarget updates an existing Webhook notification target
func (c *Client) UpdateWebhookTarget(ctx context.Context, name string, target *WebhookTarget) error {
	if name == "" {
		return fmt.Errorf("target name is required")
	}

	body := map[string]interface{}{}

	if target.URL != "" {
		body["url"] = target.URL
	}
	if target.Body != "" {
		body["body"] = target.Body
	}
	if target.Method != "" {
		body["method"] = target.Method
	}
	if target.Secret != "" {
		body["secret"] = target.Secret
	}
	if len(target.Headers) > 0 {
		body["header"] = target.Headers
	}
	if target.Comment != "" {
		body["comment"] = target.Comment
	}
	if target.Disable != nil {
		body["disable"] = *target.Disable
	}

	path := fmt.Sprintf("/config/notifications/endpoints/webhook/%s", url.PathEscape(name))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update Webhook target %s: %w", name, err)
	}

	return nil
}

// DeleteWebhookTarget deletes a Webhook notification target
func (c *Client) DeleteWebhookTarget(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("target name is required")
	}

	path := fmt.Sprintf("/config/notifications/endpoints/webhook/%s", url.PathEscape(name))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete Webhook target %s: %w", name, err)
	}

	return nil
}

// Notification Targets (Unified Listing)

// NotificationTarget represents any type of notification target (unified view from /config/notifications/targets)
type NotificationTarget struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // smtp, gotify, sendmail, webhook
	Comment string `json:"comment,omitempty"`
	Disable bool   `json:"disable,omitempty"`
	Origin  string `json:"origin,omitempty"`
}

// ListNotificationTargets lists all notification targets across all types.
// This uses the GET /config/notifications/targets API which returns a unified view
// of all targets (smtp, gotify, sendmail, webhook) configured in PBS.
func (c *Client) ListNotificationTargets(ctx context.Context) ([]NotificationTarget, error) {
	resp, err := c.api.Get(ctx, "/config/notifications/targets")
	if err != nil {
		return nil, fmt.Errorf("failed to list notification targets: %w", err)
	}

	var targets []NotificationTarget
	if err := json.Unmarshal(resp.Data, &targets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification targets: %w", err)
	}

	return targets, nil
}

// Notification Matchers

// NotificationMatcher represents a notification matcher (routing rule)
type NotificationMatcher struct {
	Name          string   `json:"name"`
	Targets       []string `json:"target,omitempty"`
	MatchSeverity []string `json:"match-severity,omitempty"` // info, notice, warning, error
	MatchField    []string `json:"match-field,omitempty"`    // field=value pairs
	MatchCalendar []string `json:"match-calendar,omitempty"` // calendar IDs
	Mode          string   `json:"mode,omitempty"`           // all, any
	InvertMatch   *bool    `json:"invert-match,omitempty"`
	Comment       string   `json:"comment,omitempty"`
	Disable       *bool    `json:"disable,omitempty"`
	Origin        string   `json:"origin,omitempty"`
	Delete        []string `json:"delete,omitempty"` // fields to delete on update
}

// ListNotificationMatchers lists all notification matchers
func (c *Client) ListNotificationMatchers(ctx context.Context) ([]NotificationMatcher, error) {
	resp, err := c.api.Get(ctx, "/config/notifications/matchers")
	if err != nil {
		return nil, fmt.Errorf("failed to list notification matchers: %w", err)
	}

	var matchers []NotificationMatcher
	if err := json.Unmarshal(resp.Data, &matchers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification matchers: %w", err)
	}

	return matchers, nil
}

// GetNotificationMatcher gets a specific notification matcher by name
func (c *Client) GetNotificationMatcher(ctx context.Context, name string) (*NotificationMatcher, error) {
	path := fmt.Sprintf("/config/notifications/matchers/%s", url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification matcher %s: %w", name, err)
	}

	var matcher NotificationMatcher
	if err := json.Unmarshal(resp.Data, &matcher); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification matcher %s: %w", name, err)
	}

	return &matcher, nil
}

// CreateNotificationMatcher creates a new notification matcher
func (c *Client) CreateNotificationMatcher(ctx context.Context, matcher *NotificationMatcher) error {
	if matcher.Name == "" {
		return fmt.Errorf("matcher name is required")
	}

	body := map[string]interface{}{
		"name": matcher.Name,
	}

	if len(matcher.Targets) > 0 {
		body["target"] = matcher.Targets
	}
	if len(matcher.MatchSeverity) > 0 {
		body["match-severity"] = matcher.MatchSeverity
	}
	if len(matcher.MatchField) > 0 {
		body["match-field"] = matcher.MatchField
	}
	if len(matcher.MatchCalendar) > 0 {
		body["match-calendar"] = matcher.MatchCalendar
	}
	if matcher.Mode != "" {
		body["mode"] = matcher.Mode
	}
	if matcher.InvertMatch != nil {
		body["invert-match"] = *matcher.InvertMatch
	}
	if matcher.Comment != "" {
		body["comment"] = matcher.Comment
	}
	if matcher.Disable != nil {
		body["disable"] = *matcher.Disable
	}

	_, err := c.api.Post(ctx, "/config/notifications/matchers", body)
	if err != nil {
		return fmt.Errorf("failed to create notification matcher %s: %w", matcher.Name, err)
	}

	return nil
}

// UpdateNotificationMatcher updates an existing notification matcher
func (c *Client) UpdateNotificationMatcher(ctx context.Context, name string, matcher *NotificationMatcher) error {
	if name == "" {
		return fmt.Errorf("matcher name is required")
	}

	body := map[string]interface{}{}

	if len(matcher.Targets) > 0 {
		body["target"] = matcher.Targets
	}
	if len(matcher.MatchSeverity) > 0 {
		body["match-severity"] = matcher.MatchSeverity
	}
	if len(matcher.MatchField) > 0 {
		body["match-field"] = matcher.MatchField
	}
	if len(matcher.MatchCalendar) > 0 {
		body["match-calendar"] = matcher.MatchCalendar
	}
	if matcher.Mode != "" {
		body["mode"] = matcher.Mode
	}
	if matcher.InvertMatch != nil {
		body["invert-match"] = *matcher.InvertMatch
	}
	if matcher.Comment != "" {
		body["comment"] = matcher.Comment
	}
	if matcher.Disable != nil {
		body["disable"] = *matcher.Disable
	}
	if len(matcher.Delete) > 0 {
		body["delete"] = matcher.Delete
	}

	path := fmt.Sprintf("/config/notifications/matchers/%s", url.PathEscape(name))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update notification matcher %s: %w", name, err)
	}

	return nil
}

// DeleteNotificationMatcher deletes a notification matcher
func (c *Client) DeleteNotificationMatcher(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("matcher name is required")
	}

	path := fmt.Sprintf("/config/notifications/matchers/%s", url.PathEscape(name))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete notification matcher %s: %w", name, err)
	}

	return nil
}
