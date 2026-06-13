package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/lobsters-cli/lobsters"
)

func (a *App) storyCmd() *cobra.Command {
	var depth int
	cmd := &cobra.Command{
		Use:   "story <short_id>",
		Short: "Fetch a story and its comment tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := lobsters.ParseShortID(args[0])
			if err != nil {
				return codeError(exitUsage, err)
			}
			a.progressf("fetching story %q (depth %d)...", id, depth)
			story, comments, err := a.client.StoryWithComments(cmd.Context(), id, depth)
			if err != nil {
				return mapFetchErr(err)
			}
			if err := a.render([]lobsters.Story{story}); err != nil {
				return err
			}
			if len(comments) > 0 {
				return a.render(comments)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&depth, "depth", -1, "comment tree depth (-1=full tree, 0=story only)")
	return cmd
}
