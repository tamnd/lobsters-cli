package lobsters

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes Lobste.rs as a kit Domain: a driver that a multi-domain host
// (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/lobsters-cli/lobsters"
//
// exactly as a database/sql program enables a driver with
// `import _ "github.com/lib/pq"`. The init below registers it; the host then
// dereferences lobsters:// URIs by routing to the operations Register installs.
// The standalone lobsters binary does not use any of this, so the CLI is
// unchanged.
func init() { kit.Register(Domain{}) }

// Domain is the Lobste.rs driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity a host reuses for help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "lobsters",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "lobsters",
			Short:  "A command line for Lobsters.",
			Long: `A command line for Lobsters (lobste.rs).

Browse hottest, newest, and active technology stories,
or filter by tag. No API key required.`,
			Site: "https://" + Host,
			Repo: "https://github.com/tamnd/lobsters-cli",
		},
	}
}

// Register installs the client factory and every Lobste.rs operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newKitClient)

	// Resolver op: fetch one story by short ID; seeds the mint map for Story.
	kit.Handle(app, kit.OpMeta{Name: "story", Group: "stories", Single: true,
		Resolver: true, URIType: "story", Summary: "Fetch a Lobsters story by short ID",
		Args: []kit.Arg{{Name: "id", Help: "story short ID"}}}, getStory)

	// List ops: story feeds.
	kit.Handle(app, kit.OpMeta{Name: "hot", Group: "stories", List: true,
		URIType: "story", Summary: "Hottest Lobsters stories"}, hotStories)

	kit.Handle(app, kit.OpMeta{Name: "newest", Group: "stories", List: true,
		URIType: "story", Summary: "Newest Lobsters stories"}, newestStories)

	kit.Handle(app, kit.OpMeta{Name: "active", Group: "stories", List: true,
		URIType: "story", Summary: "Active Lobsters stories"}, activeStories)

	kit.Handle(app, kit.OpMeta{Name: "tag", Group: "stories", List: true,
		URIType: "story", Summary: "Lobsters stories by tag",
		Args: []kit.Arg{{Name: "tag", Help: "tag name"}}}, tagStories)
}

// newKitClient builds the Lobste.rs client from the host-resolved config.
func newKitClient(_ context.Context, cfg kit.Config) (any, error) {
	lcfg := DefaultConfig()
	if cfg.UserAgent != "" {
		lcfg.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		lcfg.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		lcfg.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		lcfg.Timeout = cfg.Timeout
	}
	if cfg.Workers > 0 {
		lcfg.Workers = cfg.Workers
	}
	return NewClient(lcfg), nil
}

// storyIn is the input for fetching a single story by short ID.
type storyIn struct {
	ID     string  `kit:"arg" help:"story short ID"`
	Client *Client `kit:"inject"`
}

// listIn is the shared input for listing operations with no positional argument.
type listIn struct {
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

// tagIn is the input for the tag operation, which requires a positional tag name.
type tagIn struct {
	Tag    string  `kit:"arg" help:"tag name"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

func getStory(ctx context.Context, in storyIn, emit func(Story) error) error {
	if in.ID == "" {
		return errs.Usage("story short ID required")
	}
	story, _, err := in.Client.StoryWithComments(ctx, in.ID, 0)
	if err != nil {
		return mapKitErr(err)
	}
	return emit(story)
}

func hotStories(ctx context.Context, in listIn, emit func(Story) error) error {
	stories, err := in.Client.StoryList(ctx, "hottest", in.Limit)
	if err != nil {
		return mapKitErr(err)
	}
	for _, s := range stories {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func newestStories(ctx context.Context, in listIn, emit func(Story) error) error {
	stories, err := in.Client.StoryList(ctx, "newest", in.Limit)
	if err != nil {
		return mapKitErr(err)
	}
	for _, s := range stories {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func activeStories(ctx context.Context, in listIn, emit func(Story) error) error {
	stories, err := in.Client.StoryList(ctx, "active", in.Limit)
	if err != nil {
		return mapKitErr(err)
	}
	for _, s := range stories {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func tagStories(ctx context.Context, in tagIn, emit func(Story) error) error {
	if in.Tag == "" {
		return errs.Usage("tag name required")
	}
	stories, err := in.Client.TagStories(ctx, in.Tag, in.Limit)
	if err != nil {
		return mapKitErr(err)
	}
	for _, s := range stories {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

// mapKitErr converts a library error into the kit error kind that carries the
// right exit code.
func mapKitErr(err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case ErrRateLimited:
		return errs.RateLimited("%s", err.Error())
	case ErrNotFound:
		return errs.NotFound("%s", err.Error())
	default:
		return err
	}
}
