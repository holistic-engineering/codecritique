package critique

import (
	"context"
	"fmt"

	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type fetcher interface {
	FetchPullRequest(ctx context.Context, owner, repo, number string) (*model.PullRequest, error)
}

type reviewer interface {
	Review(context.Context, *model.PullRequest) (*model.Review, error)
}

type printer interface {
	Print(*model.Review) error
}

type Critique struct {
	fetcher  fetcher
	reviewer reviewer
	printer  printer
}

func New(
	fetcher fetcher,
	reviewer reviewer,
	printer printer,
) *Critique {
	return &Critique{
		fetcher:  fetcher,
		reviewer: reviewer,
		printer:  printer,
	}
}

func (c *Critique) Criticize(
	ctx context.Context,
	owner, repo, number string,
) error {
	// Fetch the pull request
	pr, err := c.fetcher.FetchPullRequest(ctx, owner, repo, number)
	if err != nil {
		return fmt.Errorf("failed to fetch pull request: %w", err)
	}

	// Review the pull request
	review, err := c.reviewer.Review(ctx, pr)
	if err != nil {
		return fmt.Errorf("failed to review pull request: %w", err)
	}

	// Print the review
	if err := c.printer.Print(review); err != nil {
		return fmt.Errorf("could not print review: %w", err)
	}

	return nil
}
