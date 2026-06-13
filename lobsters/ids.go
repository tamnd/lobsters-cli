package lobsters

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseShortID extracts a Lobste.rs story short ID from a bare ID string or
// a lobste.rs/s/<id> URL.
func ParseShortID(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("short_id must not be empty")
	}
	if !strings.Contains(s, "/") {
		return s, nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("not a valid short_id or URL: %q", s)
	}
	// expect path like /s/<id> or /s/<id>/title-slug
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "s" && parts[1] != "" {
		return parts[1], nil
	}
	return "", fmt.Errorf("cannot extract short_id from %q", s)
}

// ParseUsername extracts a Lobste.rs username from a bare name, a
// lobste.rs/u/<name> URL, or a lobste.rs/~<name>/... URL.
func ParseUsername(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("username must not be empty")
	}
	if !strings.Contains(s, "/") {
		return s, nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("not a valid username or URL: %q", s)
	}
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")
	// /u/<name>
	if len(parts) >= 2 && parts[0] == "u" && parts[1] != "" {
		return parts[1], nil
	}
	// /~<name> or /~<name>/...
	if len(parts) >= 1 && strings.HasPrefix(parts[0], "~") {
		name := strings.TrimPrefix(parts[0], "~")
		if name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("cannot extract username from %q", s)
}
