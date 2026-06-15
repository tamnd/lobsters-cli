// Package lobsters is the library behind the lobsters command: the HTTP client,
// request shaping, and the typed data models for Lobste.rs.
//
// The site exposes clean JSON feeds for its public content: listings, single
// stories with their comment trees, and user profiles. No API key is required.
// The Client here sets a real User-Agent, paces requests, and retries transient
// failures so a busy session stays polite.
package lobsters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Host is the canonical hostname for Lobste.rs.
const Host = "lobste.rs"

const (
	baseURL = "https://" + Host
)

// DefaultUserAgent identifies the client to Lobste.rs.
const DefaultUserAgent = "lobsters/dev (+https://github.com/tamnd/lobsters-cli)"

// ErrNotFound is returned when the API responds with 404.
var ErrNotFound = errors.New("not found")

// ErrRateLimited is returned when the API responds with 429 after all retries.
var ErrRateLimited = errors.New("rate limited")

// Config holds constructor parameters for the Client.
type Config struct {
	UserAgent string
	Rate      time.Duration
	Retries   int
	Workers   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		UserAgent: DefaultUserAgent,
		Rate:      1 * time.Second,
		Retries:   3,
		Workers:   2,
		Timeout:   30 * time.Second,
	}
}

// Client talks to Lobste.rs over HTTP.
type Client struct {
	httpClient *http.Client
	userAgent  string
	rate       time.Duration
	retries    int
	workers    int
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		userAgent:  cfg.UserAgent,
		rate:       cfg.Rate,
		retries:    cfg.Retries,
		workers:    cfg.Workers,
	}
}

// get fetches a URL with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	rateLimited := false
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, rl, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if rl {
			rateLimited = true
		}
		if !retry {
			return nil, err
		}
	}
	if rateLimited {
		return nil, ErrRateLimited
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) (body []byte, retry bool, rateLimited bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, false, ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, true, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode >= 500 {
		return nil, true, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, true, false, err
	}
	return b, false, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

// StoryList fetches a named listing feed. feed is one of "hottest", "newest",
// "active". The returned slice is capped to limit items (0 = no cap).
func (c *Client) StoryList(ctx context.Context, feed string, limit int) ([]Story, error) {
	endpoint := baseURL + "/" + feed + ".json"
	var wires []wireStory
	if err := c.getJSON(ctx, endpoint, &wires); err != nil {
		return nil, err
	}
	return wiresToStories(wires, limit), nil
}

// TagStories fetches stories for a tag. limit 0 means no cap.
func (c *Client) TagStories(ctx context.Context, tag string, limit int) ([]Story, error) {
	endpoint := baseURL + "/t/" + url.PathEscape(tag) + ".json"
	var wires []wireStory
	if err := c.getJSON(ctx, endpoint, &wires); err != nil {
		return nil, err
	}
	return wiresToStories(wires, limit), nil
}

// StoryWithComments fetches a single story and its flat comment list.
// depth -1 returns all comments; 0 returns the story with no comments;
// positive N returns comments with depth < N.
func (c *Client) StoryWithComments(ctx context.Context, shortID string, depth int) (Story, []Comment, error) {
	endpoint := baseURL + "/s/" + url.PathEscape(shortID) + ".json"
	var w wireStory
	if err := c.getJSON(ctx, endpoint, &w); err != nil {
		return Story{}, nil, fmt.Errorf("story %q: %w", shortID, err)
	}
	story := wireStoryToStory(&w)
	var comments []Comment
	if depth != 0 {
		maxDepth := depth - 1
		if depth < 0 {
			maxDepth = -1
		}
		comments = wireCommentsToComments(w.Comments, maxDepth)
	}
	return story, comments, nil
}

// User fetches a user profile.
func (c *Client) User(ctx context.Context, username string) (User, error) {
	endpoint := baseURL + "/u/" + url.PathEscape(username) + ".json"
	var w wireUser
	if err := c.getJSON(ctx, endpoint, &w); err != nil {
		return User{}, fmt.Errorf("user %q: %w", username, err)
	}
	return wireUserToUser(&w), nil
}

// UserSubmissions fetches a user's submitted stories. limit 0 means no cap.
func (c *Client) UserSubmissions(ctx context.Context, username string, limit int) ([]Story, error) {
	endpoint := baseURL + "/~" + url.PathEscape(username) + "/submitted.json"
	var wires []wireStory
	if err := c.getJSON(ctx, endpoint, &wires); err != nil {
		return nil, fmt.Errorf("submissions for %q: %w", username, err)
	}
	return wiresToStories(wires, limit), nil
}

func wiresToStories(wires []wireStory, limit int) []Story {
	out := make([]Story, 0, len(wires))
	for i := range wires {
		out = append(out, wireStoryToStory(&wires[i]))
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}
