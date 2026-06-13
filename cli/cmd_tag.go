package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) tagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag <tag>",
		Short: "Stories for a given tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tag := args[0]
			n := a.effectiveLimit(25)
			a.progressf("fetching stories for tag %q...", tag)
			stories, err := a.client.TagStories(cmd.Context(), tag, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(stories, len(stories))
		},
	}
}
