package lobsters

import (
	"strings"
	"time"
)

// Story is the record emitted for stories from any listing or single-story fetch.
type Story struct {
	ShortID      string    `json:"short_id" kit:"id"`
	ShortIDURL   string    `json:"short_id_url"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Score        int       `json:"score"`
	CommentCount int       `json:"comment_count"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	Tags         []string  `json:"tags"`
	Submitter    string    `json:"submitter"`
	SubmitterURL string    `json:"submitter_url"`
	UserIsAuthor bool      `json:"user_is_author"`
	CommentsURL  string    `json:"comments_url"`
}

// Comment is the record emitted for comments on a single story.
// The API returns comments as a flat array with a depth field already set.
type Comment struct {
	ShortID        string    `json:"short_id"`
	ShortIDURL     string    `json:"short_id_url"`
	CreatedAt      time.Time `json:"created_at"`
	LastEditedAt   time.Time `json:"last_edited_at"`
	IsDeleted      bool      `json:"is_deleted"`
	IsModerated    bool      `json:"is_moderated"`
	Score          int       `json:"score"`
	Depth          int       `json:"depth"`
	ParentComment  string    `json:"parent_comment"`
	Body           string    `json:"comment"`
	CommentingUser string    `json:"commenting_user"`
	CommentingURL  string    `json:"commenting_url"`
	URL            string    `json:"url"`
}

// User is the record emitted for a user profile.
type User struct {
	Username    string    `json:"username"`
	CreatedAt   time.Time `json:"created_at"`
	IsAdmin     bool      `json:"is_admin"`
	About       string    `json:"about"`
	IsModerator bool      `json:"is_moderator"`
	Karma       int       `json:"karma"`
	AvatarURL   string    `json:"avatar_url"`
	InvitedBy   string    `json:"invited_by_user"`
	GithubUser  string    `json:"github_username"`
	TwitterUser string    `json:"twitter_username"`
	URL         string    `json:"url"`
}

// ─── wire types ──────────────────────────────────────────────────────────────

type wireStory struct {
	ShortID       string        `json:"short_id"`
	ShortIDURL    string        `json:"short_id_url"`
	URL           string        `json:"url"`
	Title         string        `json:"title"`
	Score         int           `json:"score"`
	CommentCount  int           `json:"comment_count"`
	Description   string        `json:"description"`
	CreatedAt     time.Time     `json:"created_at"`
	Tags          []string      `json:"tags"`
	SubmitterUser string        `json:"submitter_user"`
	UserIsAuthor  bool          `json:"user_is_author"`
	CommentsURL   string        `json:"comments_url"`
	Comments      []wireComment `json:"comments"`
}

type wireComment struct {
	ShortID        string    `json:"short_id"`
	ShortIDURL     string    `json:"short_id_url"`
	CreatedAt      time.Time `json:"created_at"`
	LastEditedAt   time.Time `json:"last_edited_at"`
	IsDeleted      bool      `json:"is_deleted"`
	IsModerated    bool      `json:"is_moderated"`
	Score          int       `json:"score"`
	Depth          int       `json:"depth"`
	ParentComment  *string   `json:"parent_comment"`
	Comment        string    `json:"comment"`
	CommentingUser string    `json:"commenting_user"`
	URL            string    `json:"url"`
}

type wireUser struct {
	Username      string    `json:"username"`
	CreatedAt     time.Time `json:"created_at"`
	IsAdmin       bool      `json:"is_admin"`
	About         string    `json:"about"`
	IsModerator   bool      `json:"is_moderator"`
	Karma         int       `json:"karma"`
	AvatarURL     string    `json:"avatar_url"`
	InvitedByUser string    `json:"invited_by_user"`
	GithubUser    string    `json:"github_username"`
	TwitterUser   string    `json:"twitter_username"`
}

// ─── converters ──────────────────────────────────────────────────────────────

func wireStoryToStory(w *wireStory) Story {
	u := w.URL
	if u == "" {
		u = w.ShortIDURL
	}
	tags := w.Tags
	if tags == nil {
		tags = []string{}
	}
	return Story{
		ShortID:      w.ShortID,
		ShortIDURL:   w.ShortIDURL,
		URL:          u,
		Title:        w.Title,
		Score:        w.Score,
		CommentCount: w.CommentCount,
		Description:  strings.TrimSpace(w.Description),
		CreatedAt:    w.CreatedAt.UTC(),
		Tags:         tags,
		Submitter:    w.SubmitterUser,
		SubmitterURL: "https://lobste.rs/u/" + w.SubmitterUser,
		UserIsAuthor: w.UserIsAuthor,
		CommentsURL:  w.CommentsURL,
	}
}

func wireCommentToComment(w *wireComment) Comment {
	parent := ""
	if w.ParentComment != nil {
		parent = *w.ParentComment
	}
	return Comment{
		ShortID:        w.ShortID,
		ShortIDURL:     w.ShortIDURL,
		CreatedAt:      w.CreatedAt.UTC(),
		LastEditedAt:   w.LastEditedAt.UTC(),
		IsDeleted:      w.IsDeleted,
		IsModerated:    w.IsModerated,
		Score:          w.Score,
		Depth:          w.Depth,
		ParentComment:  parent,
		Body:           w.Comment,
		CommentingUser: w.CommentingUser,
		CommentingURL:  "https://lobste.rs/u/" + w.CommentingUser,
		URL:            w.URL,
	}
}

func wireUserToUser(w *wireUser) User {
	return User{
		Username:    w.Username,
		CreatedAt:   w.CreatedAt.UTC(),
		IsAdmin:     w.IsAdmin,
		About:       strings.TrimSpace(w.About),
		IsModerator: w.IsModerator,
		Karma:       w.Karma,
		AvatarURL:   w.AvatarURL,
		InvitedBy:   w.InvitedByUser,
		GithubUser:  w.GithubUser,
		TwitterUser: w.TwitterUser,
		URL:         "https://lobste.rs/u/" + w.Username,
	}
}

// wireCommentsToComments converts the wire flat comment slice to exported records.
// maxDepth -1 means no limit; 0 means only depth-0 comments.
func wireCommentsToComments(wcs []wireComment, maxDepth int) []Comment {
	var out []Comment
	for i := range wcs {
		wc := &wcs[i]
		if maxDepth >= 0 && wc.Depth > maxDepth {
			continue
		}
		out = append(out, wireCommentToComment(wc))
	}
	return out
}
