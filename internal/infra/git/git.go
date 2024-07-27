package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/holistic-engineering/codecritique/config"
	"github.com/holistic-engineering/codecritique/internal/critique/model"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

type Provider string

const (
	GitHub Provider = "GitHub"
	GitLab Provider = "GitLab"
)

type Client struct {
	provider     Provider
	githubClient *github.Client
	gitlabClient *gitlab.Client
}

func New(cfg *config.GitConfig) (*Client, error) {
	switch Provider(cfg.Provider) {
	case GitHub:
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cfg.Token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)
		return &Client{
			provider:     GitHub,
			githubClient: client,
		}, nil
	case GitLab:
		client, err := gitlab.NewClient(cfg.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitLab client: %w", err)
		}
		return &Client{
			provider:     GitLab,
			gitlabClient: client,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Git provider: %s", cfg.Provider)
	}
}

func (c *Client) FetchPullRequest(ctx context.Context, owner, repo, number string) (*model.PullRequest, error) {
	switch c.provider {
	case GitHub:
		return c.fetchGitHubPR(ctx, owner, repo, number)
	case GitLab:
		return c.fetchGitLabMR(owner, repo, number)
	default:
		return nil, fmt.Errorf("unsupported Git provider: %s", c.provider)
	}
}

func (c *Client) fetchGitHubPR(ctx context.Context, owner, repo, number string) (*model.PullRequest, error) {
	prNumber, err := strconv.Atoi(number)
	if err != nil {
		return nil, fmt.Errorf("invalid PR number: %w", err)
	}

	pr, _, err := c.githubClient.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub PR: %w", err)
	}

	// Fetch the diff
	opt := &github.ListOptions{
		PerPage: 100,
	}
	files, _, err := c.githubClient.PullRequests.ListFiles(ctx, owner, repo, prNumber, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub PR files: %w", err)
	}

	var diffBuilder strings.Builder
	for _, file := range files {
		diffBuilder.WriteString(fmt.Sprintf("## file: '%s'\n\n", file.GetFilename()))
		diffBuilder.WriteString(file.GetPatch())
		diffBuilder.WriteString("\n\n")
	}

	return &model.PullRequest{
		Title:       pr.GetTitle(),
		Branch:      pr.GetHead().GetRef(),
		Description: pr.GetBody(),
		Diff:        diffBuilder.String(),
	}, nil
}

func (c *Client) fetchGitLabMR(owner, repo, number string) (*model.PullRequest, error) {
	mrNumber, err := strconv.Atoi(number)
	if err != nil {
		return nil, fmt.Errorf("invalid MR number: %w", err)
	}

	mr, _, err := c.gitlabClient.MergeRequests.GetMergeRequest(owner+"/"+repo, mrNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitLab MR: %w", err)
	}

	// Fetch the diff
	changes, _, err := c.gitlabClient.MergeRequests.ListMergeRequestDiffs(owner+"/"+repo, mrNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitLab MR changes: %w", err)
	}

	var diffBuilder strings.Builder
	for _, change := range changes {
		diffBuilder.WriteString(fmt.Sprintf("## file: '%s'\n\n", change.NewPath))
		diffBuilder.WriteString(change.Diff)
		diffBuilder.WriteString("\n\n")
	}

	return &model.PullRequest{
		Title:       mr.Title,
		Branch:      mr.SourceBranch,
		Description: mr.Description,
		Diff:        diffBuilder.String(),
	}, nil
}
