package lobsters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ─── HTTP client behavior tests ───────────────────────────────────────────────

func TestGetSetsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	c := NewClient(cfg)

	_, err := c.get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.Retries = 5
	c := NewClient(cfg)

	start := time.Now()
	_, err := c.get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestGetReturnsErrNotFoundOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	c := NewClient(cfg)

	_, err := c.get(context.Background(), srv.URL)
	if err != ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// ─── wire decoding tests ──────────────────────────────────────────────────────

// The real Lobste.rs API returns submitter_user as a plain string (username),
// and comments as a flat array with a "depth" field already set.
const storyListFixture = `[
  {
    "short_id": "aaaaaa",
    "short_id_url": "https://lobste.rs/s/aaaaaa",
    "url": "https://example.com/article",
    "title": "An Example Article",
    "score": 42,
    "comment_count": 7,
    "description": "",
    "created_at": "2024-01-15T10:30:00.000-05:00",
    "tags": ["go", "programming"],
    "user_is_author": false,
    "submitter_user": "alice",
    "comments_url": "https://lobste.rs/s/aaaaaa/an_example_article"
  },
  {
    "short_id": "bbbbbb",
    "short_id_url": "https://lobste.rs/s/bbbbbb",
    "url": "",
    "title": "A Text Post",
    "score": 10,
    "comment_count": 3,
    "description": "Some markdown body here",
    "created_at": "2024-01-16T08:00:00.000Z",
    "tags": ["meta"],
    "user_is_author": true,
    "submitter_user": "bob",
    "comments_url": "https://lobste.rs/s/bbbbbb/a_text_post"
  }
]`

// Single story fixture: comments are a flat array with depth field.
const singleStoryFixture = `{
  "short_id": "aaaaaa",
  "short_id_url": "https://lobste.rs/s/aaaaaa",
  "url": "https://example.com/article",
  "title": "An Example Article",
  "score": 42,
  "comment_count": 3,
  "description": "",
  "created_at": "2024-01-15T10:30:00.000-05:00",
  "tags": ["go"],
  "user_is_author": false,
  "submitter_user": "alice",
  "comments_url": "https://lobste.rs/s/aaaaaa/an_example_article",
  "comments": [
    {
      "short_id": "c1c1c1",
      "short_id_url": "https://lobste.rs/c/c1c1c1",
      "created_at": "2024-01-15T11:00:00.000Z",
      "last_edited_at": "2024-01-15T11:00:00.000Z",
      "is_deleted": false,
      "is_moderated": false,
      "score": 5,
      "depth": 0,
      "parent_comment": null,
      "comment": "Great article!",
      "commenting_user": "carol",
      "url": "https://lobste.rs/s/aaaaaa/an_example_article#c_c1c1c1"
    },
    {
      "short_id": "c2c2c2",
      "short_id_url": "https://lobste.rs/c/c2c2c2",
      "created_at": "2024-01-15T11:30:00.000Z",
      "last_edited_at": "2024-01-15T11:30:00.000Z",
      "is_deleted": false,
      "is_moderated": false,
      "score": 2,
      "depth": 1,
      "parent_comment": "c1c1c1",
      "comment": "Agreed.",
      "commenting_user": "dave",
      "url": "https://lobste.rs/s/aaaaaa/an_example_article#c_c2c2c2"
    },
    {
      "short_id": "c3c3c3",
      "short_id_url": "https://lobste.rs/c/c3c3c3",
      "created_at": "2024-01-15T12:00:00.000Z",
      "last_edited_at": "2024-01-15T12:00:00.000Z",
      "is_deleted": true,
      "is_moderated": false,
      "score": 0,
      "depth": 0,
      "parent_comment": null,
      "comment": "",
      "commenting_user": "eve",
      "url": "https://lobste.rs/s/aaaaaa/an_example_article#c_c3c3c3"
    }
  ]
}`

const userFixture = `{
  "username": "alice",
  "created_at": "2020-06-01T00:00:00.000Z",
  "is_admin": false,
  "about": "I write Go.",
  "is_moderator": false,
  "karma": 150,
  "avatar_url": "/avatars/alice-100.png",
  "invited_by_user": "bob",
  "github_username": "alicegit",
  "twitter_username": "alicetw"
}`

func TestWireStoryToStoryList(t *testing.T) {
	var wires []wireStory
	if err := json.Unmarshal([]byte(storyListFixture), &wires); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(wires) != 2 {
		t.Fatalf("want 2 stories, got %d", len(wires))
	}

	s0 := wireStoryToStory(&wires[0])
	if s0.ShortID != "aaaaaa" {
		t.Errorf("short_id: got %q, want %q", s0.ShortID, "aaaaaa")
	}
	if s0.Title != "An Example Article" {
		t.Errorf("title: got %q", s0.Title)
	}
	if s0.Score != 42 {
		t.Errorf("score: got %d, want 42", s0.Score)
	}
	if s0.CommentCount != 7 {
		t.Errorf("comment_count: got %d, want 7", s0.CommentCount)
	}
	if s0.Submitter != "alice" {
		t.Errorf("submitter: got %q", s0.Submitter)
	}
	if s0.SubmitterURL != "https://lobste.rs/u/alice" {
		t.Errorf("submitter_url: got %q", s0.SubmitterURL)
	}
	if len(s0.Tags) != 2 || s0.Tags[0] != "go" {
		t.Errorf("tags: got %v", s0.Tags)
	}
	if s0.CreatedAt.Location() != time.UTC {
		t.Errorf("created_at not UTC: %v", s0.CreatedAt.Location())
	}
	want := time.Date(2024, 1, 15, 15, 30, 0, 0, time.UTC)
	if !s0.CreatedAt.Equal(want) {
		t.Errorf("created_at: got %v, want %v", s0.CreatedAt, want)
	}

	s1 := wireStoryToStory(&wires[1])
	// text post: url should fall back to short_id_url
	if s1.URL != "https://lobste.rs/s/bbbbbb" {
		t.Errorf("url fallback: got %q", s1.URL)
	}
	if !s1.UserIsAuthor {
		t.Errorf("user_is_author: got %v", s1.UserIsAuthor)
	}
}

func TestWireCommentsToComments(t *testing.T) {
	var w wireStory
	if err := json.Unmarshal([]byte(singleStoryFixture), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	comments := wireCommentsToComments(w.Comments, -1)
	if len(comments) != 3 {
		t.Fatalf("want 3 comments, got %d", len(comments))
	}

	// flat order: c1c1c1 (depth 0), c2c2c2 (depth 1), c3c3c3 (depth 0)
	if comments[0].ShortID != "c1c1c1" {
		t.Errorf("comments[0].short_id: got %q", comments[0].ShortID)
	}
	if comments[0].Depth != 0 {
		t.Errorf("comments[0].depth: got %d, want 0", comments[0].Depth)
	}
	if comments[1].ShortID != "c2c2c2" {
		t.Errorf("comments[1].short_id: got %q", comments[1].ShortID)
	}
	if comments[1].Depth != 1 {
		t.Errorf("comments[1].depth: got %d, want 1", comments[1].Depth)
	}
	if comments[1].ParentComment != "c1c1c1" {
		t.Errorf("comments[1].parent_comment: got %q", comments[1].ParentComment)
	}
	if comments[2].ShortID != "c3c3c3" {
		t.Errorf("comments[2].short_id: got %q", comments[2].ShortID)
	}
	if comments[2].Depth != 0 {
		t.Errorf("comments[2].depth: got %d, want 0", comments[2].Depth)
	}
	if !comments[2].IsDeleted {
		t.Errorf("comments[2].is_deleted: want true")
	}

	// commenting_url is derived
	if comments[0].CommentingURL != "https://lobste.rs/u/carol" {
		t.Errorf("comments[0].commenting_url: got %q", comments[0].CommentingURL)
	}
}

func TestWireCommentsDepthCap(t *testing.T) {
	var w wireStory
	if err := json.Unmarshal([]byte(singleStoryFixture), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// maxDepth=0 means only depth-0 comments (c1c1c1 and c3c3c3)
	top := wireCommentsToComments(w.Comments, 0)
	if len(top) != 2 {
		t.Fatalf("want 2 top-level comments with maxDepth=0, got %d", len(top))
	}
}

func TestWireUserToUser(t *testing.T) {
	var w wireUser
	if err := json.Unmarshal([]byte(userFixture), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	u := wireUserToUser(&w)

	if u.Username != "alice" {
		t.Errorf("username: got %q", u.Username)
	}
	if u.Karma != 150 {
		t.Errorf("karma: got %d, want 150", u.Karma)
	}
	if u.InvitedBy != "bob" {
		t.Errorf("invited_by_user: got %q", u.InvitedBy)
	}
	if u.GithubUser != "alicegit" {
		t.Errorf("github_username: got %q", u.GithubUser)
	}
	if u.URL != "https://lobste.rs/u/alice" {
		t.Errorf("url: got %q", u.URL)
	}
	if u.CreatedAt.Location() != time.UTC {
		t.Errorf("created_at not UTC")
	}
	wantCreated := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)
	if !u.CreatedAt.Equal(wantCreated) {
		t.Errorf("created_at: got %v, want %v", u.CreatedAt, wantCreated)
	}
}

func TestNullTagsNormalized(t *testing.T) {
	const fixture = `{
		"short_id": "zzzzzz",
		"short_id_url": "https://lobste.rs/s/zzzzzz",
		"url": "https://example.com",
		"title": "Test",
		"score": 1,
		"comment_count": 0,
		"description": "",
		"created_at": "2024-01-01T00:00:00.000Z",
		"tags": null,
		"user_is_author": false,
		"submitter_user": "x",
		"comments_url": ""
	}`
	var w wireStory
	if err := json.Unmarshal([]byte(fixture), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	s := wireStoryToStory(&w)
	if s.Tags == nil {
		t.Error("tags should be empty slice, not nil")
	}
}
