package lobsters_test

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/lobsters-cli/lobsters"
)

// These tests exercise the kit driver wiring without any network: the blank
// import side-effect of importing lobsters registers the domain, and Mint is
// pure reflection over the registry.

func TestDomainInfo(t *testing.T) {
	info := lobsters.Domain{}.Info()
	if info.Scheme != "lobsters" {
		t.Errorf("Scheme = %q, want lobsters", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != lobsters.Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, lobsters.Host)
	}
	if info.Identity.Binary != "lobsters" {
		t.Errorf("Identity.Binary = %q, want lobsters", info.Identity.Binary)
	}
}

func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := h.Domain("lobsters"); !ok {
		t.Fatal("lobsters not mounted on host")
	}

	s := lobsters.Story{ShortID: "abc123", Title: "Test Story"}
	u, err := h.Mint(s)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	want := "lobsters://story/abc123"
	if u.String() != want {
		t.Errorf("Mint = %q, want %q", u.String(), want)
	}
}
