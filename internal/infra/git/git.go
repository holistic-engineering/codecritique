package git

import (
	"context"

	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type Provider string

const (
	GitHub Provider = "GitHub"
	GitLab Provider = "GitLab"
)

type Client struct {
}

func New(provider Provider, token string) (*Client, error) {
	return &Client{}, nil
}

func (c *Client) FetchPullRequest(ctx context.Context, owner, repo, number string) (*model.PullRequest, error) {
	panic("implement me")
}
