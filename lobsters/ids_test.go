package lobsters

import "testing"

func TestParseShortIDBareName(t *testing.T) {
	id, err := ParseShortID("abcdef")
	if err != nil || id != "abcdef" {
		t.Fatalf("got (%q, %v)", id, err)
	}
}

func TestParseShortIDURL(t *testing.T) {
	id, err := ParseShortID("https://lobste.rs/s/abcdef")
	if err != nil || id != "abcdef" {
		t.Fatalf("got (%q, %v)", id, err)
	}
}

func TestParseShortIDURLWithSlug(t *testing.T) {
	id, err := ParseShortID("https://lobste.rs/s/abcdef/title-slug-here")
	if err != nil || id != "abcdef" {
		t.Fatalf("got (%q, %v)", id, err)
	}
}

func TestParseShortIDEmpty(t *testing.T) {
	_, err := ParseShortID("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestParseShortIDInvalidURL(t *testing.T) {
	// A URL with no /s/ path segment should fail
	_, err := ParseShortID("https://lobste.rs/u/someone")
	if err == nil {
		t.Fatal("expected error for non-story URL")
	}
}

func TestParseUsernameBareName(t *testing.T) {
	name, err := ParseUsername("pg")
	if err != nil || name != "pg" {
		t.Fatalf("got (%q, %v)", name, err)
	}
}

func TestParseUsernameURLSlashU(t *testing.T) {
	name, err := ParseUsername("https://lobste.rs/u/pg")
	if err != nil || name != "pg" {
		t.Fatalf("got (%q, %v)", name, err)
	}
}

func TestParseUsernameURLTilde(t *testing.T) {
	name, err := ParseUsername("https://lobste.rs/~pg/submitted.json")
	if err != nil || name != "pg" {
		t.Fatalf("got (%q, %v)", name, err)
	}
}

func TestParseUsernameEmpty(t *testing.T) {
	_, err := ParseUsername("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestParseUsernameInvalidURL(t *testing.T) {
	_, err := ParseUsername("https://lobste.rs/s/abcdef")
	if err == nil {
		t.Fatal("expected error for non-user URL")
	}
}
