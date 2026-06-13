package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) hotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hot",
		Short: "Hottest stories on Lobste.rs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(25)
			a.progressf("fetching hottest stories...")
			stories, err := a.client.StoryList(cmd.Context(), "hottest", n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(stories, len(stories))
		},
	}
}

func (a *App) newestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "newest",
		Short: "Newest stories on Lobste.rs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(25)
			a.progressf("fetching newest stories...")
			stories, err := a.client.StoryList(cmd.Context(), "newest", n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(stories, len(stories))
		},
	}
}

func (a *App) activeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "active",
		Short: "Active stories (most recently commented)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(25)
			a.progressf("fetching active stories...")
			stories, err := a.client.StoryList(cmd.Context(), "active", n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(stories, len(stories))
		},
	}
}
