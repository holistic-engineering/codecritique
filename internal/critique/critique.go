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

type Critique struct {
	fetcher  fetcher
	reviewer reviewer
}

func New(fetcher fetcher, reviewer reviewer) *Critique {
	return &Critique{
		fetcher:  fetcher,
		reviewer: reviewer,
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

	// Print the review results
	fmt.Printf("Review for PR #%s in %s/%s\n", number, owner, repo)
	fmt.Printf("Title: %s\n", review.PullRequest.Title)
	fmt.Printf("Description: %s\n\n", review.PullRequest.Description)

	for _, suggestion := range review.Suggestions {
		fmt.Printf("File: %s\n", suggestion.File.Name)
		fmt.Printf("Line: %d\n", suggestion.Line)
		fmt.Printf("Suggestion: %s\n\n", suggestion.Message)
	}

	return nil
}
