package ai

import (
	"context"

	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type Kind string

const (
	KindChatGPT4o      Kind = ""
	KindClaude35Sonnet Kind = ""
)

type Client struct {
}

func New(kind Kind) *Client {
	return &Client{}
}

func (c *Client) Review(context.Context, *model.PullRequest) (*model.Review, error) {
	panic("implement me")
}
