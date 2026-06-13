package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/lobsters-cli/lobsters"
)

func (a *App) userCmd() *cobra.Command {
	var submissions bool
	cmd := &cobra.Command{
		Use:   "user <username>",
		Short: "Show a Lobste.rs user profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := lobsters.ParseUsername(args[0])
			if err != nil {
				return codeError(exitUsage, err)
			}
			a.progressf("fetching user %q...", name)
			user, err := a.client.User(cmd.Context(), name)
			if err != nil {
				return mapFetchErr(err)
			}
			if err := a.render([]lobsters.User{user}); err != nil {
				return err
			}
			if submissions {
				n := a.effectiveLimit(25)
				a.progressf("fetching submissions for %q (limit %d)...", name, n)
				stories, err := a.client.UserSubmissions(cmd.Context(), name, n)
				if err != nil {
					return mapFetchErr(err)
				}
				return a.render(stories)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&submissions, "submissions", false, "also list the user's submitted stories")
	return cmd
}
