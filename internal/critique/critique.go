package critique

import (
	"context"

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
	// TODO: implement logic
	return nil
}
