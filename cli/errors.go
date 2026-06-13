package cli

import (
	"errors"

	"github.com/tamnd/lobsters-cli/lobsters"
)

func isNotFound(err error) bool {
	return errors.Is(err, lobsters.ErrNotFound)
}

func isRateLimited(err error) bool {
	return errors.Is(err, lobsters.ErrRateLimited)
}
